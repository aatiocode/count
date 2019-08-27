package repository

import (
	"database/sql"
	"fmt"
	"github.com/aristio/count/campaigns"
	"github.com/aristio/count/models"
	"github.com/aristio/count/rewards"
	"time"

	"github.com/labstack/echo"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type psqlCampaignRepository struct {
	Conn    *sql.DB
	rwdRepo rewards.Repository
}

// NewPsqlCampaignRepository will create an object that represent the campaigns.Repository interface
func NewPsqlCampaignRepository(Conn *sql.DB, rwdRepo rewards.Repository) campaigns.Repository {
	return &psqlCampaignRepository{Conn, rwdRepo}
}

func (m *psqlCampaignRepository) CreateCampaign(c echo.Context, campaign *models.Campaign) error {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	now := time.Now()
	query := `INSERT INTO campaigns (name, description, start_date, end_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	var lastID int64

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	err = stmt.QueryRow(campaign.Name, campaign.Description, campaign.StartDate, campaign.EndDate,
		campaign.Status, &now).Scan(&lastID)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	campaign.ID = lastID
	campaign.CreatedAt = &now
	return nil
}

func (m *psqlCampaignRepository) UpdateCampaign(c echo.Context, id int64,
	updateCampaign *models.Campaign) error {

	var lastID int64
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	now := time.Now()
	query := `UPDATE campaigns SET status = $1, updated_at = $2 WHERE id = $3 RETURNING id`
	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	err = stmt.QueryRow(updateCampaign.Status, &now, id).Scan(&lastID)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	return nil
}

func (m *psqlCampaignRepository) UpdateExpiryDate(c echo.Context) error {
	var lastID int64
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	now := time.Now()
	query := `UPDATE campaigns SET status = 0, updated_at = $1
		WHERE end_date::timestamp::date < now()::date AND status = 1`
	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		requestLogger.Debug("Update Status Base on Expiry Date: ", err)

		return err
	}

	err = stmt.QueryRow(&now).Scan(&lastID)

	if err != nil {
		requestLogger.Debug("Update Status Base on Expiry Date: ", err)

		return err
	}

	return nil
}

func (m *psqlCampaignRepository) UpdateStatusBasedOnStartDate() error {
	var lastID int64
	now := time.Now()
	query := `UPDATE campaigns SET status = 1, updated_at = $1
		WHERE start_date::timestamp::date = now()::date`
	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		logrus.Debug("Update Status Base on Start Date: ", err)

		return err
	}

	err = stmt.QueryRow(&now).Scan(&lastID)

	if err != nil {
		logrus.Debug("Update Status Base on Start Date: ", err)

		return err
	}

	return nil
}

func (m *psqlCampaignRepository) GetCampaign(c echo.Context, payload map[string]interface{}) ([]*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	paging := ""
	where := ""
	query := `SELECT id, name, description, start_date, end_date, status, updated_at, created_at,
		DATE_PART('day', end_date::timestamp - now()::timestamp) as days_remaining
		FROM campaigns WHERE id IS NOT NULL`

	if payload["page"].(int) > 0 || payload["limit"].(int) > 0 {
		paging = fmt.Sprintf(" LIMIT %d OFFSET %d",
			payload["limit"].(int), ((payload["page"].(int) - 1) * payload["limit"].(int)))
	}

	if payload["name"].(string) != "" {
		where += " AND name LIKE '%" + payload["name"].(string) + "%'"
	}

	if payload["status"].(string) != "" {
		where += " AND status='" + payload["status"].(string) + "'"
	}

	if payload["startDate"].(string) != "" {
		where += " AND start_date >= '" + payload["startDate"].(string) + "'"
	}

	if payload["endDate"].(string) != "" {
		where += " AND end_date <= '" + payload["endDate"].(string) + "'"
	}

	query += where + " ORDER BY status DESC, days_remaining ASC" + paging
	res, err := m.getCampaign(c, query)

	if err != nil {
		requestLogger.Debug(err)

		return nil, err
	}

	return res, err

}

func (m *psqlCampaignRepository) getCampaign(c echo.Context, query string) ([]*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	result := make([]*models.Campaign, 0)

	rows, err := m.Conn.Query(query)

	if err != nil {
		requestLogger.Debug(err)

		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		t := new(models.Campaign)
		var createDate, updateDate pq.NullTime
		var daysRemaining int64

		err = rows.Scan(
			&t.ID,
			&t.Name,
			&t.Description,
			&t.StartDate,
			&t.EndDate,
			&t.Status,
			&updateDate,
			&createDate,
			&daysRemaining,
		)

		if err != nil {
			requestLogger.Debug(err)

			return nil, err
		}

		t.CreatedAt = &createDate.Time
		t.UpdatedAt = &updateDate.Time

		// get rewards
		rewards, err := m.rwdRepo.GetRewardByCampaign(c, t.ID)

		if err != nil {
			requestLogger.Debug(err)

			return nil, err
		}

		t.Rewards = &rewards

		result = append(result, t)
	}

	return result, nil
}

func (m *psqlCampaignRepository) GetCampaignAvailable(c echo.Context, today string) ([]*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)

	query := fmt.Sprintf(`SELECT id, name, description, start_date, end_date, status, updated_at,
		created_at, DATE_PART('day', end_date::timestamp - now()::timestamp) as days_remaining
		FROM campaigns WHERE status = 1 AND start_date::date <= '%s'
		AND end_date::date >= '%s' ORDER BY start_date DESC`, today, today)

	res, err := m.getCampaign(c, query)

	if err != nil {
		requestLogger.Debug(err)

		return nil, err
	}

	return res, err
}

func (m *psqlCampaignRepository) SavePoint(c echo.Context, cmpgnTrx *models.CampaignTrx) error {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	now := time.Now()
	var id int64
	var query string

	if cmpgnTrx.TransactionType == models.TransactionPointTypeDebet {
		query = `INSERT INTO campaign_transactions (user_id, point_amount, transaction_type, transaction_date, reff_core, campaign_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)  RETURNING id`
		id = cmpgnTrx.Campaign.ID
	}

	if cmpgnTrx.TransactionType == models.TransactionPointTypeKredit {
		query = `INSERT INTO campaign_transactions (user_id, point_amount, transaction_type, transaction_date, reff_core, voucher_code_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)  RETURNING id`
		id = cmpgnTrx.VoucherCode.ID
	}

	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	cmpgnTrx.CreatedAt = &now
	var lastID int64
	err = stmt.QueryRow(cmpgnTrx.CIF, *cmpgnTrx.PointAmount, cmpgnTrx.TransactionType, cmpgnTrx.TransactionDate, cmpgnTrx.RefCore, id, cmpgnTrx.CreatedAt).Scan(&lastID)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	cmpgnTrx.ID = lastID
	return nil
}

func (m *psqlCampaignRepository) CountCampaign(c echo.Context, payload map[string]interface{}) (int, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	var total int
	query := `SELECT coalesce(COUNT(id), 0) FROM campaigns WHERE id IS NOT NULL`
	where := ""

	if payload["name"].(string) != "" {
		where += " AND name LIKE '%" + payload["name"].(string) + "%'"
	}

	if payload["status"].(string) != "" {
		where += " AND status='" + payload["status"].(string) + "'"
	}

	if payload["startDate"].(string) != "" {
		where += " AND start_date >= '" + payload["startDate"].(string) + "'"
	}

	if payload["endDate"].(string) != "" {
		where += " AND end_date <= '" + payload["endDate"].(string) + "'"
	}

	query += where
	err := m.Conn.QueryRow(query).Scan(&total)

	if err != nil {
		requestLogger.Debug(err)

		return 0, err
	}

	return total, nil
}

func (m *psqlCampaignRepository) GetCampaignDetail(c echo.Context, id int64) (*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	var createDate, updateDate pq.NullTime
	result := new(models.Campaign)

	query := `SELECT id, name, description, start_date, end_date, status, updated_at, created_at FROM campaigns WHERE id = $1`

	err := m.Conn.QueryRow(query, id).Scan(
		&result.ID,
		&result.Name,
		&result.Description,
		&result.StartDate,
		&result.EndDate,
		&result.Status,
		&createDate,
		&updateDate,
	)

	if err != nil {
		requestLogger.Debug(err)

		return nil, err
	}

	result.CreatedAt = &createDate.Time
	result.UpdatedAt = &updateDate.Time

	return result, nil
}

func (m *psqlCampaignRepository) Delete(c echo.Context, id int64) error {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	query := `DELETE FROM campaigns WHERE ID = $1`
	stmt, err := m.Conn.Prepare(query)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	result, err := stmt.Query(id)

	if err != nil {
		requestLogger.Debug(err)

		return err
	}

	defer result.Close()
	requestLogger.Debug("Result delete campaign: ", result)

	return nil
}
