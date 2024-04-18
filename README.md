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
# Брокер сообщений
Используется ```Rabbit MQ 3.10.7```, хотя в начале я хотел использовать ```Apache Kafka```. Со второй у меня возникли трудности в настройке, если конкретно, то я не смог добиться равномерного распределения задач по вычислителям. В брокере создается две очереди: в первую оркестратор пишет задачи, которые получают вычислители, а во вторую вычислители пишут ответы, которые получает оркестратор. 
# Вычислитель
Скопирован из предфинальной задачи. Парсер, который я написал самостоятельно, малофункциональный и не самый оптимальный. Вычислительный сервер запускает указанное количество вычислителей. Из за парсера вычислитель умеет считать только выражения из целых положительных чисел без скобок. Поддерживается только сложение, вычитание, деление и умножение.

Принцип деления выражения на подзадачи: Возьмем выражение 1*2+2-3*4/7-9, в нем имеют приоритет операции 1*2 и 3*4/7, то есть группы в котрых только умножения и деления, такие группы будут запущены в отдельных горутинах, когда все группы будут подсчитаны, можно выпонять операции второго приоритета, то есть начнется вычисление выражения 2+2-1.714285-9. Если умножение выполняется за 10 секунд, деление за 2 секунды, сложение за 5, а вычитание за 1, то такое выражение будет подсчитано за 12+5+1+1=19 секунды, так как 1*2 и 3*4/7 считаются параллельно, но 3*4/7 считается на две секунды дольше, итого 12.

Я хотел написать новый парсер (Честно), но к сожалению из за учебы времени не хватило.
# Запуск и тестирование
Для корректного запуска лучше остановить все работающие контейнеры, так как они могут занимать такие же порты:
```
docker stop $(docker ps -a -q)
```
Все упаковано в docker-compose. Для запуска в Linux нужжно переидти в директорию проекта и ввести команду:
```
docker-compose up
```
Для остановки:
```
docker-compose down
```
Важно: команда ```docker-compose down``` удалит контейнеры, то есть данные в базе данных потеряются. Для остановки без удаления используйте запуск с логами с последующим отклбчением через Ctrl+C в терминале.

Установить докер:
```
curl -sSL https://get.docker.com | sh
```
```
sudo usermod -aG docker $(whoami)
```

Для просмотра веб страницы переидите по: http://localhost:8081/registration
Так же можно зайти в базу данных в рабочем docker-сонтейнере:
```
docker exec -it leonids_postgres psql -U leonid -W postgres
```
Для тестирования отказов можно остановить один из работающих контейнеров
Посмотреть список работающик контейнеров:
```
docker ps
```
Остановить контейнер:
```
docker stop <id>
```
# Проблемы и недоработки на текущий коммит(если успею пофиксить, это раздет уменьшится)
1. JWT-токен работает 15 минут. По канону нужно делать два токена, один обновляется часто, а второй используется для обновления первого и обновляется только повторным вводом логина и пароля. Я решил пока оставить упрощенный вариант с одним токеном. Если 15 минут пройдет, придется возвращаться на предыдущую страницу и логиниться снова.
2. Приложение переживает перезапуск вычислителя, базы данных, оркестратора и фронта, однако если упадет Кролик то подключение с ним восстановлено не будет.
3. На четвертой вкладке, где происходит мониторинг вычислителей, каждый пользователь открыто видит какие выражения считают вычислители в текущий момент. Такое поведение нежелательно, так как система все таки для разных пользователей, и нехорошо что один пользователь может какое то время наблюдать выражение отправленное другим пользователем. 

Тому кто это проверяет: 
Фууух) Вот и подощло финальное задание к концу. Все мы очень хорошо постарались, путь был долгий и не простой, но мы как то дотянули до финиша. Код мой неидеальный, есть недоработки и недочеты, но не судите очень строго :)
По вопросам пиши мне в телеграм: https://t.me/QwErTy256fsoceity
Я догадываюсь что многие работают под виндой, а я пишу терминальные команды для линукса, если что то не работает из за ОС то обращайся, разберемся.










