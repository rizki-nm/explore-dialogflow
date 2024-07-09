package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	KomplainPesanan   = "1komplain-pesanan-sent-intent"
	KonfirmasiPesanan = "2konfirmasi-pesanan-sent-intent"

	KomplainPesananFallback   = "1komplain-pesanan-fallback-intent"
	KonfirmasiPesananFallback = "2konfirmasi-pesanan-fallback-intent"
)

type Pesanan struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

type DialogFlowWebhookResponse struct {
	FollowupEventInput struct {
		Name         string `json:"name"`
		LanguageCode string `json:"languageCode"`
		Parameters   struct {
			ParamName string `json:"param-name"`
		} `json:"parameters"`
	} `json:"followupEventInput"`
}

func extractCustomerData(data string) (int, string, error) {
	// Split the input data by spaces
	parts := strings.Fields(data)

	// Ensure the data format is as expected
	if len(parts) < 4 || parts[0] != "Nomor" || parts[1] != "ID:" || parts[3] != "Nama:" {
		return 0, "", errors.New("invalid data format")
	}

	// Extract customer ID
	idStr := parts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, "", err
	}

	// Extract customer name
	nama := parts[4]

	return id, nama, nil
}

func processIntent(ctx context.Context, data string, intent string) DialogFlowWebhookResponse {
	pesanans := []Pesanan{
		{ID: 1234, Nama: "Joko"},
		{ID: 4567, Nama: "Budi"},
		{ID: 6789, Nama: "Susi"},
	}

	fallbackIntentName := ""
	if intent == KomplainPesanan {
		fallbackIntentName = KomplainPesananFallback
	} else if intent == KonfirmasiPesanan {
		fallbackIntentName = KonfirmasiPesananFallback
	}

	exist := false

	customerID, customerName, err := extractCustomerData(data)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed to extract customer data: %v", err)

		response := DialogFlowWebhookResponse{}
		response.FollowupEventInput.Name = fallbackIntentName
		response.FollowupEventInput.LanguageCode = "en-US"

		return response
	}

	for _, pesanan := range pesanans {
		if (intent == KomplainPesanan || intent == KonfirmasiPesanan) && pesanan.ID == customerID && pesanan.Nama == customerName {
			exist = true
			break
		}
	}

	if !exist {
		log.Ctx(ctx).Error().Msgf("Customer data not available for ID: %d, Name: %s", customerID, customerName)

		response := DialogFlowWebhookResponse{}
		response.FollowupEventInput.Name = fallbackIntentName
		response.FollowupEventInput.LanguageCode = "en-US"

		log.Ctx(ctx).Debug().Msgf("Response: %s", response)

		return response
	}

	response := DialogFlowWebhookResponse{}
	response.FollowupEventInput.Name = "handover-intent"
	response.FollowupEventInput.LanguageCode = "en-US"

	return response
}

type DialogflowWebhookRequest struct {
	QueryResult struct {
		QueryText string `json:"queryText"`
		Intent    struct {
			DisplayName string `json:"displayName"`
		} `json:"intent"`
	} `json:"queryResult"`
	Session string `json:"session"`
}

func (d *DialogflowWebhookRequest) Intent() string {
	return d.QueryResult.Intent.DisplayName
}

func (d *DialogflowWebhookRequest) QueryText() string {
	return d.QueryResult.QueryText
}

func dialogflowWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var req DialogflowWebhookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fulfillmentResp := processIntent(r.Context(), req.QueryText(), req.Intent())

	log.Ctx(r.Context()).Debug().Msgf("FullfilmentResp: %s", fulfillmentResp)

	resp, _ := json.Marshal(fulfillmentResp)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	r := http.NewServeMux()
	r.Handle("GET /", http.HandlerFunc(healthcheckHandler))
	r.Handle("POST /wh/dialogflow", http.HandlerFunc(dialogflowWebhookHandler))

	h := chainMiddleware(
		r,
		recoverHandler,
		loggerHandler(func(w http.ResponseWriter, r *http.Request) bool { return r.URL.Path == "/" }),
		realIPHandler,
		requestIDHandler,
	)

	srv := http.Server{
		Addr:         ":8080",
		Handler:      h,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	go func() {
		log.Info().Msgf("server is starting on port 8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal().Msgf("could not listen server: %s", err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)

	log.Info().Msgf("shutting down")
	os.Exit(0)
}
