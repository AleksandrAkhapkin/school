package infrastruct

import "net/http"

type CustomError struct {
	msg  string
	Code int
}

func NewError(msg string, code int) *CustomError {
	return &CustomError{
		msg:  msg,
		Code: code,
	}
}

func (c *CustomError) Error() string {
	return c.msg
}

var (
	ErrorEmailIsExist        = NewError("email уже зарегистрирован", http.StatusBadRequest)
	ErrorInternalServerError = NewError("внутренняя ошибка сервера", http.StatusInternalServerError)
	ErrorBadRequest          = NewError("плохие входные данные запроса", http.StatusBadRequest)
	ErrorJWTIsBroken         = NewError("jwt испорчен", http.StatusForbidden)
	ErrorPermissionDenied    = NewError("у вас недостаточно прав", http.StatusForbidden)
	ErrorPasswordIsIncorrect = NewError("неверный пароль", http.StatusForbidden)
	ErrorPasswordsDoNotMatch = NewError("пароли не совпадают", http.StatusBadRequest)
	ErrorEmailNotFind        = NewError("Пользователь с таким Email не найден", http.StatusBadRequest)

	ErrorNotFound = NewError("материалы не найдены", http.StatusNotFound)
)
