package business

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// WageManager calculates and pays worker wages during settlement.
type WageManager struct {
	queries *sqlc.Queries
	ledger  *economy.Ledger
	log     zerolog.Logger
}

// NewWageManager initializes the wage manager.
func NewWageManager(q *sqlc.Queries, l *economy.Ledger, log zerolog.Logger) *WageManager {
	return &WageManager{
		queries: q,
		ledger:  l,
		log:     log.With().Str("component", "business_wages").Logger(),
	}
}

// PayWages calculates and distributes wages for all active employment relationships.
func (w *WageManager) PayWages(ctx context.Context) error {
	log := logger.FromContext(ctx, "business.wages")
	log.Info().Msg("starting daily wage distribution")

	businesses, err := w.queries.GetActiveBusinesses(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active businesses: %w", err)
	}

	totalPaid := decimal.Zero
	totalDefaulted := decimal.Zero

	for _, biz := range businesses {
		paid, defaulted, err := w.processBusinessWages(ctx, biz)
		if err != nil {
			log.Error().Err(err).Str("business_id", biz.ID).Msg("failed to process business wages")
			continue
		}
		totalPaid = totalPaid.Add(paid)
		totalDefaulted = totalDefaulted.Add(defaulted)
	}

	log.Info().
		Str("total_paid", totalPaid.String()).
		Str("total_defaulted", totalDefaulted.String()).
		Msg("daily wage distribution completed")

	return nil
}

func (w *WageManager) processBusinessWages(ctx context.Context, biz sqlc.Business) (decimal.Decimal, decimal.Decimal, error) {
	log := logger.FromContext(ctx, "business.wages")

	// 1. Fetch all active employees for this business
	employees, err := w.queries.GetEmploymentByBusiness(ctx, biz.ID)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("failed to fetch employees: %w", err)
	}

	if len(employees) == 0 {
		return decimal.Zero, decimal.Zero, nil
	}

	// 2. Fetch the business's operating account (fallback to owner's wallet if no dedicated business account)
	bizAccount, err := w.queries.GetBusinessAccount(ctx, biz.ID)
	if err != nil {
		bizAccount, err = w.queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
			PlayerID:    biz.OwnerID,
			AccountType: sqlc.AccountTypeWALLET,
		})
		if err != nil {
			return decimal.Zero, decimal.Zero, fmt.Errorf("failed to fetch business/owner account: %w", err)
		}
	}

	var totalPaid decimal.Decimal
	var totalDefaulted decimal.Decimal

	// 3. Iterate through employees and pay them
	for _, emp := range employees {
		labor, err := w.queries.GetDailyLaborByEmployeeAndBusiness(ctx, sqlc.GetDailyLaborByEmployeeAndBusinessParams{
			PlayerID:   emp.EmployeeID,
			BusinessID: biz.ID,
		})

		var hoursWorked decimal.Decimal
		if err == nil {
			hoursWorked = numericToDecimal(labor.HoursWorked)
		} else {
			// If no labor logged, assume minimum daily hours for salaried workers
			hoursWorked = numericToDecimal(emp.MinDailyHours)
		}

		wageRate := numericToDecimal(emp.WageRate)
		totalWage := wageRate.Mul(hoursWorked).Truncate(0)

		if totalWage.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// Fetch employee's wallet
		empWallet, err := w.queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
			PlayerID:    emp.EmployeeID,
			AccountType: sqlc.AccountTypeWALLET,
		})
		if err != nil {
			log.Warn().Err(err).Str("employee_id", emp.EmployeeID).Msg("failed to fetch employee wallet, skipping")
			continue
		}

		// Execute transfer via Ledger (Atomic Double-Entry)
		err = w.ledger.Transfer(ctx, bizAccount.ID, empWallet.ID, totalWage, "WAGE_PAYMENT", emp.EmployeeID, fmt.Sprintf("Wages from %s", biz.Name))
		if err != nil {
			if errors.Is(err, economy.ErrInsufficientFunds) {
				log.Warn().
					Str("business_id", biz.ID).
					Str("employee_id", emp.EmployeeID).
					Str("amount", totalWage.String()).
					Msg("business lacks funds to pay employee, debt accrued")
				
				// Layer 4+: Set debt flag on business or log to a debts table
				totalDefaulted = totalDefaulted.Add(totalWage)
				continue
			}
			return totalPaid, totalDefaulted, fmt.Errorf("ledger transfer failed for employee %s: %w", emp.EmployeeID, err)
		}

		totalPaid = totalPaid.Add(totalWage)
	}

	// 4. Clear daily labor records for this business to prepare for the next day
	err = w.queries.ClearDailyLaborForBusiness(ctx, biz.ID)
	if err != nil {
		log.Warn().Err(err).Str("business_id", biz.ID).Msg("failed to clear daily labor records")
	}

	return totalPaid, totalDefaulted, nil
}

// numericToDecimal safely converts pgtype.Numeric to decimal.Decimal.
func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}