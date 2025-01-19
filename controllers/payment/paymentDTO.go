package controllers

import (
	payment "gambl/core/payment"
)

// Request DTOs

type CreatePaymentLinkDTO struct {
	payment.CreatePaymentLink
}
