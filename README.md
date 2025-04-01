# diploma

ответ api
"INFO"/"ERROR": code, message
обязательно нужно обрабатывать ответы с плохим кодом

есть кука "userRole" - "student"/"teacher"
она ставится после авторизации

страница авторизации
endpoint - "/"
две кнопки - "войти" и "зарегистрироваться", каждая кнопка открывает свою страничку

страница входа
при нажатии отправляет post на "/api/login" с зашифрованным паролем и логином
при успехе открывается главная страница

страница регистрации
при нажатии оправляет post на "/api/register" с выбранной ролью, зашифрованным паролем и логином
при успехе открывается страница заполнения профиля

страница заполнения профиля
для студента: заполнить поля фамилия, имя, отчество, возраст
для преподавателя: заполнить поля фамилия, имя, отчество, возраст, цена, специализация
фио пускай заполняют в разных полях, а отправляется одной строкой fio
при нажатии отправляется post на "/api/fill-profile" с заполненными данными и открывается главная страница

главная страница
для преподавателя выводится меню с опциями: мои студенты, мои задания, мой профиль
мои студенты - get на "/api/get-students", в ответ json список usersList без цены и специальности, ID выводить не надо
у каждого студента две кнопки: добавить задание и начать чат
добавить задание: открывается проводник для выбора docx файла и при успешном выборе оправляется запрос upload как в task.html
начать чат: пока кнока заглушка
type usersList struct {
	ID        uuid.UUID `json:"id"`
	Fio       string    `json:"fio"`
	Age       uint8     `json:"age"`
	Specialty string    `json:"specialty,omitempty"`
	Price     int       `json:"price,omitempty"`
	Rating    float32   `json:"rating"`
}
мои задания - get на "/api/get-tasks", в ответ json список taskList, ID выводить не надо, если поля grade нет, выводить оценку не надо
type taskList struct {
	ID     uuid.UUID `json:"taskID"`
	Name   string    `json:"taskName"`
	Grade  uint8     `json:"grade,omitempty"`
	Status string    `json:"status"`
}
мой профиль - get на "/api/get-profile", в ответ одна запись userList, ID выводить не надо

для студента выводится меню с опциями: все преподаватели, мои преподаватели, мои задания, мой профиль
все преподаватели - get на "/api/get-teachers" в ответ json список usersList, ID выводить не надо
также есть кнопка выбора сортировки по возрастанию/убыванию и выбора поля для сортировки по рейтингу/цене
у каждого преподавателя одна кнопка: отправить заявку
отправить заявку: post на "/api/send-request" с id преподавателя
мои преподаватели - get на "/api/get-my-teachers", в ответ json список usersList, ID выводить не надо
у каждого преподавателя одна кнопка: начать чат
начать чат: пока кнока заглушка
type usersList struct {
	ID        uuid.UUID `json:"id"`
	Fio       string    `json:"fio"`
	Age       uint8     `json:"age"`
	Specialty string    `json:"specialty,omitempty"`
	Price     int       `json:"price,omitempty"`
	Rating    float32   `json:"rating"`
}
мои задания - get на "/api/get-tasks", в ответ json список taskList, ID выводить не надо, если поля grade нет, выводить оценку не надо
type taskList struct {
	ID     uuid.UUID `json:"taskID"`
	Name   string    `json:"taskName"`
	Grade  uint8     `json:"grade,omitempty"`
	Status string    `json:"status"`
}
мой профиль - get на "/api/get-profile", в ответ одна запись userList, ID выводить не надо