package loyalityservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/go-resty/resty/v2"
)

func worker(ctx context.Context, id int, orderIn chan entities.Order, saveOrderOut chan entities.Order, accrualURL string, errorChan chan error) {
	preffix := fmt.Sprintf("worker #%d", id)
	for {
		select {
		case <-ctx.Done():
			errorChan <- makeWorkerError(preffix, errors.New("stopped by context"))
			return
		case order := <-orderIn:
			var accrualOrder AccrualOrder
			response, err := resty.New().R().SetResult(&accrualOrder).Get(accrualURL + "/api/orders/" + order.Number)
			if err != nil {
				errorChan <- makeWorkerError(preffix, err)
				continue
			}

			switch response.StatusCode() {
			case http.StatusNoContent:
				order.Status = entities.OrderStatusInvalid
			case http.StatusTooManyRequests:
				return
			case http.StatusInternalServerError:
				return
			}

			if response.StatusCode() == http.StatusOK {

				if order.Number != accrualOrder.Order {
					errorChan <- makeWorkerError(preffix, errors.New("wrong order number"))
					continue
				}

				switch accrualOrder.Status {
				case entities.AccrealStatusRegistered:
					continue
				case entities.AccrealStatusProcessing:
					order.Status = entities.OrderStatusProcessing
				case entities.AccrealStatusInvalid:
					order.Status = entities.OrderStatusInvalid
				case entities.AccrealStatusProcessed:
					order.Status = entities.OrderStatusProcessed
					order.Accrual = accrualOrder.Accrual
				}
			}

			order.UpdatedAt = time.Now()
			saveOrderOut <- order
		}
	}
}

func makeWorkerError(preffix string, err error) error {
	return fmt.Errorf("%s: %s", preffix, err.Error())
}
