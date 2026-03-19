package service

import (
	"context"
	"log"
	"time"

	"banking-service/internal/client"
	"banking-service/internal/model"
	"banking-service/internal/repository"
)

const retryAfter = 72 * time.Hour

type LoanScheduler struct {
	loanRepo    repository.LoanRepository
	accountRepo repository.AccountRepository
	txRepo      repository.TransactionRepository
	txProcessor *TransactionProcessor
	mailer      Mailer
	userClient  client.UserClient
	loanSvc     *LoanService
}

func NewLoanScheduler(
	loanRepo repository.LoanRepository,
	accountRepo repository.AccountRepository,
	txRepo repository.TransactionRepository,
	txProcessor *TransactionProcessor,
	mailer Mailer,
	userClient client.UserClient,
	loanSvc *LoanService,
) *LoanScheduler {
	return &LoanScheduler{
		loanRepo:    loanRepo,
		accountRepo: accountRepo,
		txRepo:      txRepo,
		txProcessor: txProcessor,
		mailer:      mailer,
		userClient:  userClient,
		loanSvc:     loanSvc,
	}
}

// Start pokrece dve goroutine: dnevni cron za naplatu rata i mesecni cron
// za azuriranje varijabilnih kamatnih stopa
func (s *LoanScheduler) Start(ctx context.Context) {
	go s.runDailyInstallmentJob(ctx)
	go s.runMonthlyRateAdjustmentJob(ctx)
}

func (s *LoanScheduler) runDailyInstallmentJob(ctx context.Context) {
	// pokrecemo odmah pri startu, a zatim cekamo sledeci dan u ponoc
	s.processDueInstallments(ctx)
	s.processRetryInstallments(ctx)

	for {
		timer := time.NewTimer(time.Until(nextMidnight()))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.processDueInstallments(ctx)
			s.processRetryInstallments(ctx)
		}
	}
}

// processDueInstallments trazi sve PENDING rate ciji je DueDate danas ili ranije
func (s *LoanScheduler) processDueInstallments(ctx context.Context) {
	now := time.Now()
	log.Printf("[LoanScheduler] processDueInstallments started at %s", now.Format(time.RFC3339))

	installments, err := s.loanRepo.FindDueInstallments(ctx, now)
	if err != nil {
		log.Printf("[LoanScheduler] FindDueInstallments error: %v", err)
		return
	}

	for i := range installments {
		s.processInstallment(ctx, &installments[i])
	}

	log.Printf("[LoanScheduler] processDueInstallments done, processed %d installments", len(installments))
}

// processRetryInstallments trazi RETRYING rate kojima je dosao retry_at
func (s *LoanScheduler) processRetryInstallments(ctx context.Context) {
	now := time.Now()
	log.Printf("[LoanScheduler] processRetryInstallments started at %s", now.Format(time.RFC3339))

	installments, err := s.loanRepo.FindRetryInstallments(ctx, now)
	if err != nil {
		log.Printf("[LoanScheduler] FindRetryInstallments error: %v", err)
		return
	}

	for i := range installments {
		s.processInstallment(ctx, &installments[i])
	}

	log.Printf("[LoanScheduler] processRetryInstallments done, processed %d installments", len(installments))
}

// processInstallment pokusava naplatu jedne rate
func (s *LoanScheduler) processInstallment(ctx context.Context, installment *model.LoanInstallment) {
	loan := &installment.Loan

	account, err := s.accountRepo.FindByAccountNumber(ctx, loan.AccountNumber)
	if err != nil || account == nil {
		log.Printf("[LoanScheduler] account not found for loan %d: %v", loan.ID, err)
		return
	}

	// proveravamo da li ima dovoljno sredstava
	if account.AvailableBalance < installment.Amount {
		log.Printf("[LoanScheduler] installment %d payment failed: insufficient funds", installment.ID)
		s.onInstallmentFailed(ctx, installment, loan)
		return
	}

	bankAccountNumber, ok := BankAccounts[account.Currency.Code]
	if !ok {
		log.Printf("[LoanScheduler] no bank account for currency %s", account.Currency.Code)
		return
	}

	bankAccount, err := s.accountRepo.FindByAccountNumber(ctx, bankAccountNumber)
	if err != nil || bankAccount == nil {
		log.Printf("[LoanScheduler] bank account not found: %v", err)
		return
	}

	// kreiramo transaction record
	transaction := &model.Transaction{
		PayerAccountNumber:     loan.AccountNumber,
		RecipientAccountNumber: bankAccountNumber,
		StartAmount:            installment.Amount,
		StartCurrencyCode:      account.Currency.Code,
		EndAmount:              installment.Amount,
		EndCurrencyCode:        account.Currency.Code,
		Status:                 model.TransactionCompleted,
	}

	if err := s.txRepo.Create(ctx, transaction); err != nil {
		log.Printf("[LoanScheduler] failed to create transaction for installment %d: %v", installment.ID, err)
		return
	}

	// direktno azuriramo balanse
	model.UpdateBalances(account, -installment.Amount)
	model.UpdateBalances(bankAccount, installment.Amount)

	if err := s.accountRepo.UpdateBalance(ctx, account); err != nil {
		log.Printf("[LoanScheduler] failed to update client balance for installment %d: %v", installment.ID, err)
		return
	}
	if err := s.accountRepo.UpdateBalance(ctx, bankAccount); err != nil {
		log.Printf("[LoanScheduler] failed to update bank balance for installment %d: %v", installment.ID, err)
		return
	}

	s.onInstallmentPaid(ctx, installment, loan)
}
func (s *LoanScheduler) onInstallmentPaid(ctx context.Context, installment *model.LoanInstallment, loan *model.Loan) {
	now := time.Now()
	installment.Status = model.InstallmentStatusPaid
	installment.PaidAt = &now
	installment.RetryAt = nil

	if err := s.loanRepo.UpdateInstallment(ctx, installment); err != nil {
		log.Printf("[LoanScheduler] UpdateInstallment (paid) error: %v", err)
		return
	}

	loan.RemainingDebt -= installment.Amount
	if loan.RemainingDebt < 0 {
		loan.RemainingDebt = 0
	}
	loan.PaidInstallments++

	// ako je ovo bila poslednja rata, zatvaramo kredit
	if loan.PaidInstallments >= loan.RepaymentPeriod {
		loan.Status = model.LoanStatusCompleted
		loan.NextInstallmentDate = time.Time{}
	} else {
		loan.NextInstallmentDate = loan.NextInstallmentDate.AddDate(0, 1, 0)
	}

	if err := s.loanRepo.UpdateLoan(ctx, loan); err != nil {
		log.Printf("[LoanScheduler] UpdateLoan (paid) error: %v", err)
	}

	log.Printf("[LoanScheduler] installment %d paid, loan %d remaining debt: %.2f", installment.ID, loan.ID, loan.RemainingDebt)
}

// onInstallmentFailed postavlja retry ili trajno UNPAID i salje notifikaciju klijentu
func (s *LoanScheduler) onInstallmentFailed(ctx context.Context, installment *model.LoanInstallment, loan *model.Loan) {
	// ako je vec bila RETRYING, oznacavamo kao trajno neplacenu
	if installment.Status == model.InstallmentStatusRetrying {
		installment.Status = model.InstallmentStatusUnpaid
		installment.RetryAt = nil
		log.Printf("[LoanScheduler] installment %d permanently UNPAID after retry", installment.ID)
	} else {
		retryAt := time.Now().Add(retryAfter)
		installment.Status = model.InstallmentStatusRetrying
		installment.RetryAt = &retryAt
		log.Printf("[LoanScheduler] installment %d set to RETRYING at %s", installment.ID, retryAt.Format(time.RFC3339))
	}

	if err := s.loanRepo.UpdateInstallment(ctx, installment); err != nil {
		log.Printf("[LoanScheduler] UpdateInstallment (failed) error: %v", err)
		return
	}

	s.sendFailureNotification(ctx, loan, installment)
}

// sendFailureNotification dohvata email klijenta iz user-service-a i salje obavestenje
func (s *LoanScheduler) sendFailureNotification(ctx context.Context, loan *model.Loan, installment *model.LoanInstallment) {
	clientResp, err := s.userClient.GetClientByID(ctx, loan.ClientID)
	if err != nil {
		log.Printf("[LoanScheduler] GetClientByID failed for client %d: %v", loan.ClientID, err)
		return
	}

	subject := "Neuspesna naplata rate kredita"
	body := "Naplata rate vašeg kredita nije bila uspešna zbog nedovoljnih sredstava na računu. " +
		"Molimo vas da obezbedite sredstva. Novi pokušaj će biti izvršen u roku od 72 sata."

	// ako je retry vec bio neuspesan, saljemo ozbiljnije obavestenje
	if installment.Status == model.InstallmentStatusUnpaid {
		subject = "Rata kredita nije naplaćena"
		body = "Uprkos ponovljenom pokušaju, naplata rate vašeg kredita nije bila uspešna. " +
			"Molimo vas da kontaktirate banku."
	}

	if err := s.mailer.Send(clientResp.Email, subject, body); err != nil {
		log.Printf("[LoanScheduler] failed to send failure notification for loan %d: %v", loan.ID, err)
	}
}

func (s *LoanScheduler) runMonthlyRateAdjustmentJob(ctx context.Context) {
	for {
		timer := time.NewTimer(time.Until(nextFirstOfMonth()))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			log.Printf("[LoanScheduler] runMonthlyRateAdjustmentJob started")
			if err := s.loanSvc.AdjustVariableRates(ctx); err != nil {
				log.Printf("[LoanScheduler] AdjustVariableRates error: %v", err)
			}
		}
	}
}

// nextMidnight vraca vreme sledece ponoci (00:00:00 sutra)
func nextMidnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

// nextFirstOfMonth vraca 1. u sledecem mesecu u 01:00
func nextFirstOfMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month()+1, 1, 1, 0, 0, 0, now.Location())
}
