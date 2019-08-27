package main

import (
	"database/sql"
	"fmt"
	"github.com/aristio/count/campaigns"
	"github.com/aristio/count/middleware"
	"github.com/aristio/count/models"
	"github.com/aristio/count/services"
	"github.com/aristio/count/vouchers"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	_campaignHttpDelivery "github.com/aristio/count/campaigns/delivery/http"
	_campaignRepository "github.com/aristio/count/campaigns/repository"
	_campaignUseCase "github.com/aristio/count/campaigns/usecase"

	"github.com/carlescere/scheduler"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var ech *echo.Echo
var metricService services.MetricService

func init() {
	ech = echo.New()
	ech.Debug = true
	loadEnv()
	logrus.SetReportCaller(true)
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: models.DateTimeFormatMillisecond + "000",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			tmp := strings.Split(f.File, "/")
			filename := tmp[len(tmp)-1]
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
	}

	logrus.SetFormatter(formatter)
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	dbConn := getDBConn()
	migrate := dataMigrations(dbConn)

	defer dbConn.Close()
	defer migrate.Close()

	contextTimeout, err := strconv.Atoi(os.Getenv(`CONTEXT_TIMEOUT`))

	if err != nil {
		fmt.Println(err)
	}

	timeoutContext := time.Duration(contextTimeout) * time.Second

	echoGroup := models.EchoGroup{
		Admin: ech.Group("/admin"),
		API:   ech.Group("/api"),
		Token: ech.Group("/token"),
	}

	// load all middlewares
	middleware.InitMiddleware(ech, echoGroup)

	// PING
	ech.GET("/ping", ping)

	// TOKEN
	tokenRepository := _tokenRepository.NewPsqlTokenRepository(dbConn)
	tokenUseCase := _tokenUseCase.NewTokenUseCase(tokenRepository, timeoutContext)
	_tokenHttpDelivery.NewTokensHandler(echoGroup, tokenUseCase)

	// TAG
	tagRepository := _tagRepository.NewPsqlTagRepository(dbConn)
	tagUseCase := _tagUseCase.NewTagUseCase(tagRepository)

	// POINTHISTORY
	pHistoryRepository := _pHistoryRepository.NewPsqlPointHistoryRepository(dbConn)
	pHistoryUseCase := _pHistoryUseCase.NewPointHistoryUseCase(pHistoryRepository)
	_pHistoryHttpDelivery.NewPointHistoriesHandler(echoGroup, pHistoryUseCase)

	// REWARDTRX
	rewardTrxRepository := _rewardTrxRepository.NewPsqlRewardTrxRepository(dbConn)
	rewardTrxUseCase := _rewardTrxUC.NewRewardtrxUseCase(rewardTrxRepository)
	_rewardTrxHttpDelivery.NewRewardTrxHandler(echoGroup, rewardTrxUseCase)

	// QUOTA
	quotaRepository := _quotaRepository.NewPsqlQuotaRepository(dbConn)
	quotaUseCase := _quotaUseCase.NewQuotaUseCase(quotaRepository, rewardTrxUseCase)

	// VOUCHER
	voucherRepository := _voucherRepository.NewPsqlVoucherRepository(dbConn)

	// VOUCHERCODE
	voucherCodeRepository := _voucherCodeRepository.NewPsqlVoucherCodeRepository(dbConn)
	voucherCodeUseCase := _voucherCodeUseCase.NewVoucherCodeUseCase(voucherCodeRepository, voucherRepository)
	_voucherCodeHttpDelivery.NewVoucherCodesHandler(echoGroup, voucherCodeUseCase)

	// REWARD
	rewardRepository := _rewardRepository.NewPsqlRewardRepository(dbConn, quotaRepository, tagRepository)
	campaignRepository := _campaignRepository.NewPsqlCampaignRepository(dbConn, rewardRepository)
	voucherUseCase := _voucherUseCase.NewVoucherUseCase(voucherRepository, campaignRepository, pHistoryRepository)
	_voucherHttpDelivery.NewVouchersHandler(echoGroup, voucherUseCase)
	rewardUseCase := _rewardUseCase.NewRewardUseCase(rewardRepository, campaignRepository, tagUseCase, quotaUseCase, voucherUseCase, voucherCodeRepository, rewardTrxRepository)
	_rewardHttpDelivery.NewRewardHandler(echoGroup, rewardUseCase)

	// CAMPAIGN
	campaignUseCase := _campaignUseCase.NewCampaignUseCase(campaignRepository, rewardUseCase)
	_campaignHttpDelivery.NewCampaignsHandler(echoGroup, campaignUseCase)

	// CAMPAIGNTRX
	campaignTrxRepository := _campaignTrxRepository.NewPsqlCampaignTrxRepository(dbConn)
	campaignTrxUseCase := _campaignTrxUseCase.NewCampaignTrxUseCase(campaignTrxRepository)
	_campaignTrxHttpDelivery.NewCampaignTrxsHandler(echoGroup, campaignTrxUseCase, campaignUseCase)

	// USER
	userRepository := _userRepository.NewPsqlUserRepository(dbConn)
	userUseCase := _userUseCase.NewUserUseCase(userRepository, timeoutContext)
	_userHttpDelivery.NewUserHandler(echoGroup, userUseCase)

	// METRIC
	metricRepository := _metricRepository.NewPsqlMetricRepository(dbConn)
	metricUseCase := _metricUseCase.NewMetricUseCase(metricRepository, timeoutContext)

	// Add metric
	_metricService.NewMetricHandler(metricUseCase)

	// Run every day.
	updateStatusBasedOnStartDate(campaignUseCase, voucherUseCase)

	// run refresh reward trx
	rewardUseCase.RefreshTrx()

	ech.Start(":" + os.Getenv(`PORT`))

}

func updateStatusBasedOnStartDate(cmp campaigns.UseCase, vcr vouchers.UseCase) {
	scheduler.Every().Day().At(os.Getenv(`STATUS_UPDATE_TIME`)).Run(func() {
		t := time.Now()
		logrus.Debug("Run Scheduler! @", t)

		// CAMPAIGN
		cmp.UpdateStatusBasedOnStartDate()

		// VOUCHER
		vcr.UpdateStatusBasedOnStartDate()
	})
}

func ping(echTx echo.Context) error {
	res := echTx.Response()
	rid := res.Header().Get(echo.HeaderXRequestID)
	params := map[string]interface{}{"rid": rid}

	requestLogger := logrus.WithFields(logrus.Fields{"params": params})

	requestLogger.Info("Start to ping server.")
	response := models.Response{}
	response.Status = models.StatusSuccess
	response.Message = "PONG!!"

	requestLogger.Info("End of ping server.")

	return echTx.JSON(http.StatusOK, response)
}

func getDBConn() *sql.DB {
	dbHost := os.Getenv(`DB_HOST`)
	dbPort := os.Getenv(`DB_PORT`)
	dbUser := os.Getenv(`DB_USER`)
	dbPass := os.Getenv(`DB_PASS`)
	dbName := os.Getenv(`DB_NAME`)

	connection := fmt.Sprintf("postgres://%s%s@%s%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	dbConn, err := sql.Open(`postgres`, connection)

	if err != nil {
		logrus.Debug(err)
	}

	err = dbConn.Ping()

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	return dbConn
}

func dataMigrations(dbConn *sql.DB) *migrate.Migrate {
	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})

	migrations, err := migrate.NewWithDatabaseInstance(
		"file://migrations/",
		os.Getenv(`DB_USER`), driver)

	if err != nil {
		logrus.Debug(err)
	}

	if err := migrations.Up(); err != nil {
		logrus.Debug(err)
	}

	return migrations
}

func loadEnv() {
	// check .env file existence
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return
	}

	err := godotenv.Load()

	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	return
}
