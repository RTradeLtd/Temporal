package v2

import (
	"github.com/RTradeLtd/Temporal/queue"
)

// stripeTemplate is used to render a checkout button
// allowing purchasing of credits using credit cards
var stripeTemplate = `<html>
<head>
  <title>{{ .title }}</title>
</head>
<body>
<form action="/v2/stripe/charge" method="post" class="payment">
  <input type="hidden" value="{{ .amount }}" name="provided_amount">
  <script src="https://checkout.stripe.com/checkout.js" class="stripe-button"
	data-key="{{ .Key }}"
	data-name="Temporal"
    data-description="{{ .description }}"
	data-amount="{{ .amount }}"
	data-email="{{ .email }}"
	data-zip-code="true"
	data-billing-address="true"
	data-allow-remember-me="false"
    data-locale="auto"></script>
</form>
</body>
</html>`

// CreditRefund is a data object to contain refund information
type CreditRefund struct {
	Username string
	CallType string
	Cost     float64
}

type queues struct {
	pin      *queue.Manager
	cluster  *queue.Manager
	email    *queue.Manager
	ipns     *queue.Manager
	key      *queue.Manager
	database *queue.Manager
	dash     *queue.Manager
	eth      *queue.Manager
}
