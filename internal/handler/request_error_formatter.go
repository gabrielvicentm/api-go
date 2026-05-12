package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func formatRequestError(err error) string {
	if err == nil {
		return ""
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		messages := make([]string, 0, len(validationErrors))
		for _, fieldErr := range validationErrors {
			fieldName := humanizeRequestFieldName(fieldErr.Field())
			messages = append(messages, fmt.Sprintf("%s %s", fieldName, translateValidationTag(fieldErr.Tag())))
		}

		return strings.Join(messages, " ")
	}

	var unmarshalTypeError *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeError) {
		fieldName := humanizeRequestFieldName(unmarshalTypeError.Field)
		return fmt.Sprintf("%s deve ser do tipo %s.", fieldName, humanizeRequestType(unmarshalTypeError.Type.String()))
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return "JSON invalido no corpo da requisicao."
	}

	return strings.TrimSpace(err.Error())
}

func humanizeRequestFieldName(field string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return "Campo"
	}

	field = strings.ReplaceAll(field, ".", " ")
	field = strings.ReplaceAll(field, "_", " ")

	var builder strings.Builder
	for index, char := range field {
		if index > 0 && char >= 'A' && char <= 'Z' && field[index-1] != ' ' {
			builder.WriteByte(' ')
		}

		builder.WriteRune(char)
	}

	words := strings.Fields(builder.String())
	if len(words) == 0 {
		return "Campo"
	}

	for index, word := range words {
		upperWord := strings.ToUpper(word)
		switch upperWord {
		case "CPF", "CNPJ", "CEP", "CNH", "RG", "UF", "ID", "KM", "PIX":
			words[index] = upperWord
		default:
			words[index] = strings.ToLower(word)
		}
	}

	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]
	return strings.Join(words, " ")
}

func translateValidationTag(tag string) string {
	switch tag {
	case "required":
		return "e obrigatorio."
	case "email":
		return "deve ser um e-mail valido."
	case "min":
		return "esta abaixo do minimo permitido."
	case "max":
		return "ultrapassa o maximo permitido."
	case "len":
		return "deve ter o tamanho esperado."
	case "uuid":
		return "deve ser um identificador valido."
	case "oneof":
		return "tem um valor invalido."
	default:
		return fmt.Sprintf("falhou na validacao %q.", tag)
	}
}

func humanizeRequestType(typeName string) string {
	switch typeName {
	case "string":
		return "texto"
	case "float64", "float32", "int", "int32", "int64", "uint", "uint32", "uint64":
		return "numero"
	case "bool":
		return "booleano"
	default:
		return typeName
	}
}
