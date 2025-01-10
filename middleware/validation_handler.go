package middleware

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Validator instance untuk digunakan dalam validasi
var validate *validator.Validate

func init() {
	// Sinkronkan validator bawaan Gin dengan custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate = v

		// Register custom validation rule
		validate.RegisterValidation("name_format", func(fl validator.FieldLevel) bool {
			re := regexp.MustCompile(`^[A-Za-z\s]+$`)
			return re.MatchString(fl.Field().String())
		})
		// Register custom validation rule untuk format tanggal (hanya tanggal tanpa jam)
		validate.RegisterValidation("date_format", func(fl validator.FieldLevel) bool {
			// Format yang digunakan untuk validasi tanggal: YYYY-MM-DD
			_, err := time.Parse("2006-01-02", fl.Field().String())
			return err == nil
		})
	}
}

// Daftar tabel schema
type TransactionSchema struct {
	UserID   uint   `json:"user_id" binding:"omitempty,numeric"`
	ItemID   uint   `json:"item_id" binding:"required,numeric"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Status   string `json:"status" binding:"omitempty"`
}

type DetailSchema struct {
	Code   string `json:"code" binding:"omitempty"`
	Out    string `json:"out" binding:"omitempty,date_format"`
	Entry  string `json:"entry" binding:"omitempty,date_format"`
	Status string `json:"status" binding:"omitempty"`
}

type ItemSchema struct {
	Name  string `json:"name" binding:"required,name_format"`
	Stock int    `json:"stock" binding:"required,min=0"`
}

type UpdateSchema struct {
	Name     string `json:"name" binding:"omitempty,name_format"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=3"`
}

type UserSchema struct {
	Name     string `json:"name" binding:"required,name_format"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=3"`
}

type LoginSchema struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=3"`
}

// FormatValidationErrors mengembalikan semua pesan kesalahan validasi sebagai array string
func FormatValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			message := formatFieldError(fieldError)
			errors = append(errors, message)
		}
	} else {
		errors = append(errors, err.Error())
	}
	return errors
}

// formatFieldError memformat pesan error untuk satu field dengan pesan khusus
func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "name_format":
		return fmt.Sprintf("Field '%s' must not contain symbols or numbers.", fe.Field())
	case "required":
		return fmt.Sprintf("Field '%s' must be filled in.", fe.Field())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address.", fe.Field())
	case "min":
		return fmt.Sprintf("Field '%s' must have at least %s characters.", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("Field '%s' invalid: %s", fe.Field(), fe.Tag())
	}
}

// ValidateInput memvalidasi input berdasarkan skema yang diberikan
func ValidateInput(schema interface{}) []string {
	err := validate.Struct(schema)
	if err != nil {
		return FormatValidationErrors(err)
	}
	return nil
}
