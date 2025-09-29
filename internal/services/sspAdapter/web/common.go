package sppAdapterWeb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ggicci/httpin"
	"github.com/redis/go-redis/v9"
	"gitlab.com/twinbid-exchange/RTB-exchange/internal/constants"
	utils "gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/utils_grpc"
)

func getNurl(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	timeout time.Duration,
) {
	input := r.Context().Value(httpin.Input).(*nurlRequest)

	decodedURL, err := url.QueryUnescape(input.DspURL)
	if err != nil {
		log.Printf("Failed to decode original URL: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		client := &http.Client{Timeout: timeout}
		resp, err := client.Get(decodedURL)
		if err != nil {
			log.Printf("Failed to proxy win notice to DSP %s, globalID: %s, error: %w",
				decodedURL,
				input.GlobalId,
				err,
			)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			log.Printf("DSP %s returned error for win notice: %d", decodedURL, resp.StatusCode)
			w.WriteHeader(resp.StatusCode)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func getBurl(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	redisClient *redis.Client,
	timeout time.Duration,
) {
	input := r.Context().Value(httpin.Input).(*burlRequest)

	decodedURL, err := url.QueryUnescape(input.DspURL)
	if err != nil {
		log.Printf("Failed to decode original URL: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	} else {
		client := &http.Client{Timeout: timeout}
		resp, err := client.Get(decodedURL)
		if err != nil {
			log.Printf(
				"Failed to proxy billable event to DSP %s, globalID: %s, error: %w",
				decodedURL,
				input.GlobalId,
				err,
			)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			log.Printf("DSP %s returned error for billable event: %d", decodedURL, resp.StatusCode)
			w.WriteHeader(resp.StatusCode)
		}
	}

	if err := utils.WriteStringToRedis(ctx, redisClient, input.GlobalId, constants.RESULT_COLUMN, constants.SUCCESS); err != nil {
		fmt.Printf("failed to WriteStringToRedis SUCCESS in getBurl: %w", err)
	}

	w.WriteHeader(http.StatusOK)
}

func getHealth(
	w http.ResponseWriter,
) {
	w.WriteHeader(http.StatusNoContent)
}
