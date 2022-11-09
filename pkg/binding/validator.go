package binding

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	vEn "github.com/go-playground/validator/v10/translations/en"
	"github.com/pariz/gountries"

	"github.com/coretrix/hitrix/pkg/errors"
	"github.com/coretrix/hitrix/service"
)

var (
	once      sync.Once
	singleton *Validator
)

type Validator struct {
	validator  *validator.Validate
	translator ut.Translator
}

func (t *Validator) ValidateStruct(s interface{}) error {
	err := t.validator.Struct(s)
	if err != nil {
		var fieldErrors errors.FieldErrors = make(map[string]string)

		validatorErrs := err.(validator.ValidationErrors)
		for _, e := range validatorErrs {
			translatedErr := e.Translate(t.translator)
			fieldErrors[e.Field()] = translatedErr
		}

		return fieldErrors
	}

	return nil
}

func (t *Validator) Engine() interface{} {
	return t.validator
}

func (t *Validator) Validate(field interface{}, rules string) []error {
	err := t.validator.Var(field, rules)

	return t.translateError(err)
}

func NewValidator() *Validator {
	once.Do(func() {
		validatorInstance := validator.New()
		validatorInstance.SetTagName("binding")
		english := en.New()
		uni := ut.New(english, english)
		translator, _ := uni.GetTranslator("en")

		for _, customValidation := range customValidations {
			err := validatorInstance.RegisterValidation(customValidation.RuleName, customValidation.ValidatorFunction)
			if err != nil {
				panic(err)
			}

			err = validatorInstance.RegisterTranslation(
				customValidation.RuleName,
				translator, customValidation.CustomRegisFunc,
				customValidation.CustomTransFunc,
			)
			if err != nil {
				panic(err)
			}
		}
		_ = vEn.RegisterDefaultTranslations(validatorInstance, translator)

		singleton = &Validator{validator: validatorInstance, translator: translator}
	})

	return singleton
}

func (t *Validator) translateError(err error) (errs []error) {
	if err == nil {
		return nil
	}

	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(t.translator))
		errs = append(errs, translatedErr)
	}

	return errs
}

type CustomValidation struct {
	RuleName          string
	ValidatorFunction func(validator.FieldLevel) bool
	CustomRegisFunc   validator.RegisterTranslationsFunc
	CustomTransFunc   validator.TranslationFunc
}

var customValidations = []CustomValidation{
	{
		RuleName:          "timestamp_gte_now",
		ValidatorFunction: validateTimestampGteNow,
		CustomRegisFunc: func(ut ut.Translator) error {
			return ut.Add("timestamp_gte_now", "value should be greater than now", true)
		},
		CustomTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("timestamp_gte_now", fe.Field())

			return t
		},
	},
	{
		RuleName:          "country_code_custom",
		ValidatorFunction: validateCountryCodeAlpha2,
		CustomRegisFunc: func(ut ut.Translator) error {
			return ut.Add("country_code_custom", "not a valid Country Code", true)
		},
		CustomTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("country_code_custom", fe.Field())

			return t
		},
	},
	{
		RuleName:          "password_strength",
		ValidatorFunction: validatePasswordStrength(8),
		CustomRegisFunc: func(ut ut.Translator) error {
			return ut.Add(
				"password_strength",
				"Not strong enough. Should be more than 8 character, contain at least 1 lowercase, 1 uppercase, 1 number, and 1 special character.",
				false)
		},
		CustomTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("password_strength", fe.Field())

			return t
		},
	},
}

func validateCountryCodeAlpha2(fl validator.FieldLevel) bool {
	query := gountries.New()
	_, err := query.FindCountryByAlpha(fl.Field().String())

	return err == nil
}

func validateTimestampGteNow(fl validator.FieldLevel) bool {
	value := time.UnixMilli(fl.Field().Int())

	var today time.Time
	if service.HasService(service.ClockService) {
		today = service.DI().Clock().Now()
	} else {
		today = time.Now().Truncate(time.Second)
	}

	return today.Before(value)
}

func validatePasswordStrength(minLength int) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		pass := fl.Field().String()

		if len(pass) < minLength {
			return false
		}

		ok, _ := regexp.MatchString(`[a-z]+`, pass)
		if !ok {
			return false
		}

		ok, _ = regexp.MatchString(`[A-Z]+`, pass)
		if !ok {
			return false
		}

		ok, _ = regexp.MatchString(`[0-9]+`, pass)
		if !ok {
			return false
		}

		// ref: https://owasp.org/www-community/password-special-characters
		specialChars := " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
		ok, _ = regexp.MatchString("["+specialChars+"]+", pass)

		return ok
	}
}
