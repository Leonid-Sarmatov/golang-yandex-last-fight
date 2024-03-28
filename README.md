# Выпускной проект от яндекс лицея
По функционалу является расширением для предыдущего проекта, добавляется авторизация, добавляются новые требования. Ссылка на предфинальную задачу: https://github.com/Leonid-Sarmatov/golang-yandex-final-boss
# Техническое задание
Продолжаем работу над проектом Распределенный калькулятор
В этой части работы над проектом реализуем персистентность (возможность программы восстанавливать свое состояние после перезагрузки) и многопользовательский режим.
Простыми словами: все, что мы делали до этого теперь будет работать в контексте пользователей, а все данные будут храниться в СУБД

Функционал
Добавляем регистрацию пользователя
Пользователь отправляет запрос
POST /api/v1/register {
"login": ,
"password":
}
В ответ получае 200+OK (в случае успеха). В противном случае - ошибка
Добавляем вход
Пользователь отправляет запрос
POST /api/v1/login {
"login": ,
"password":
}
В ответ получае 200+OK и JWT токен для последующей авторизации.

Весь реализованный ранее функционал работает как раньше, только в контексте конкретного пользователя.
У кого выражения хранились в памяти - переводим хранение в SQLite. (теперь наша система обязана переживать перезагрузку)
У кого общение вычислителя и сервера вычислений было реализовано с помощью HTTP - переводим взаимодействие на GRPC
