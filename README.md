# Выпускной проект от яндекс лицея
По функционалу является расширением для предыдущего проекта, добавляется авторизация, добавляются новые требования. Ссылка на предфинальную задачу: https://github.com/Leonid-Sarmatov/golang-yandex-final-boss
# Техническое задание
Продолжаем работу над проектом Распределенный калькулятор
В этой части работы над проектом реализуем персистентность (возможность программы восстанавливать свое состояние после перезагрузки) и многопользовательский режим.
Простыми словами: все, что мы делали до этого теперь будет работать в контексте пользователей, а все данные будут храниться в СУБД

Функционал
Добавляем регистрацию пользователя
Пользователь отправляет запрос
```
POST /api/v1/register {
"login": ,
"password":
}
```
В ответ получае 200+OK (в случае успеха). В противном случае - ошибка
Добавляем вход
Пользователь отправляет запрос
```
POST /api/v1/login {
"login": ,
"password":
}
```
В ответ получае 200+OK и JWT токен для последующей авторизации.

Весь реализованный ранее функционал работает как раньше, только в контексте конкретного пользователя.
У кого выражения хранились в памяти - переводим хранение в SQLite. (теперь наша система обязана переживать перезагрузку)
У кого общение вычислителя и сервера вычислений было реализовано с помощью HTTP - переводим взаимодействие на GRPC
# Описание принципа работы
Вся система состоит из нескольких частей:
 - Фронтэнд
 - Оркестратор
 - База данных
 - Брокер соощений
 - Вычислительный сервер
# Фронтэнд 
Состоит из простого сервера, поднимающего две html-разметки. Первая представляет собой страницу регистрации и входа в аккаунт. Вторая - эта страница аккаунта. При успешной аутентификации оркестратор возвращает JWT-токен, которых сохраняется в сессии браузера, а затем пользователь автоматически перенапрявляется на страницу аккаунта. Теоретически зная эндпоинт для страницы аккаунта попасть на нее можно, однако, так как нет JWT-токена, авторизации не будет, что дает защиту данных пользователя.
# Оркестратор
Представляет собой API предоставляющее ендпоинты для HTTP запросов клиента. Обмен данными происходит в формате JSON. 
Эндпоинты:
 - ```/login```, принимает запрос для аутентификации
 - ```/registration```, принимает запрос для регистрации нового пользователя
 - ```/api/sendTask```, принимает запрос с задачей, и возвращает ошибку если задача невалидна
 - ```/api/getListOfTask```, возвращает список со всеми задачами
 - ```/api/sendTimeOfOperations```, принимает время выполнения для каждой операции
 - ```/api/getListOfSolvers```, возвращает список с доступными вычислителями

Оркестратор написан при помощи роутера CHI. Во время старта оркестратор инициализирует таблицы в базе данных, на случай если их нет, инициализирует подключение к брокеру сообщений, а так же и инициализирует GRPC сервер, принимаеющий запросы сердцебиения вычислителей. Оркестратор регулярно проверяет время с последнего седцебиения, и если оно слишком велико, то сервер отмечает вычислителя как "мертвого". Для авторизации все запросы начинающиеся с ```/api``` проходят middleware, проверяющий валидность токена. Все задачи, а так же время выполнения сохраняются в базу данных. Для распределения задач по вычислителям используется брокер сообщений, который настроен так, что бы равномерно распределять задачи по доступным вычислителям. 

Папка ```handlers``` содержит пакеты с функциями, которые возвращают соответствующий хэндлер. Эти фукции принимают на вход объекты структур, которые имплиментируют интерфейсы необходимые для работы конкретного хэндлера. Такой подход позволяет проще менять или дополнять функционал приложения, например в будущем можно будет добавить кеширование данных и заменить Postgress на что то иное, главное что бы новые структуры реализовывали нужные методы.
# База данных
Используется ```PostgreSQL 16.1```. В базе данных созданы три таблицы:
 - Таблица пользователей
 - Таблица задач
 - Таблица времени выполнения

Пароли пользователей хранятся в виде хешей, что позволяет повысить безопасность при захвате злоумышленником базы данных. Каждой задаче, а так же времени выполнения присваивается имя пользователя, в результате чего пользователи могут использовать приложение независимо друг от друга.










