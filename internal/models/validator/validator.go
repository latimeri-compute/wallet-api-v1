package validator

import "slices"

type ValidatorInterface interface {
	Check(checkOk bool, key, message string)
	Valid() bool
}

type Validator struct {
	Errors map[string]string
}

// создаёт новый валидатор
func NewValidator() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// возвращает true если validator.errors пуст
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// добавляет ошибку в Validator.errors в формате "поле": "возникшая ошибка"
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

/*
добавляет ошибку в Validator.errors если значение checkOk != true

пример: validator.Check(EmptyString != "", StringFieldKey, "поле не может быть пустым")
*/
func (v *Validator) Check(checkOk bool, key, message string) {
	if !checkOk {
		v.AddError(key, message)
	}
}

// возвращает true, если значение value есть в списке permittedValues
func IsPermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
