package common

// IDParam is a reusable struct for parsing an :id path param.
type IDParam struct {
	ID int `params:"id" validate:"required,gt=0"`
}
