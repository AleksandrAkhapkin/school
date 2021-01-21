package postgres

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tarasova-school/internal/types"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "err with Open DB")
	}

	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "err with ping DB")
	}

	return &Postgres{db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) CreateUser(user *types.User) (int, error) {
	var id int
	if err := p.db.QueryRow("INSERT INTO users (email, pass, first_name, user_role) VALUES ($1, $2, $3, $4)"+
		" RETURNING id", user.Email, user.Password, user.FirstName, user.UserRole).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) CreateTeacher(teacher *types.Teacher) error {
	tx, err := p.db.Begin()
	if err != nil {
		return errors.Wrap(err, "err with Begin")
	}

	err = tx.QueryRow("INSERT INTO users (email, pass, first_name, user_role) VALUES ($1, $2, $3, $4) "+
		"RETURNING id", teacher.Email, teacher.Password, teacher.FirstName, teacher.UserRole).Scan(&teacher.ID)
	if err != nil {
		return errors.Wrap(err, "err with Postgress bd users")
	}

	_, err = tx.Exec("INSERT INTO teacher_info (id) VALUES ($1)",
		teacher.ID)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "err with Postgress bd teacher_info")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "err with Commit")
	}
	return nil
}

func (p *Postgres) GetTeacherByID(id int) (*types.TeacherFullInfo, error) {

	teacher := types.TeacherFullInfo{ID: id}
	err := p.db.QueryRow("SELECT users.first_name, teacher_info.good, teacher_info.improve, "+
		"teacher_info.ahtung, users.times_seconds FROM users, teacher_info WHERE "+
		"users.id = teacher_info.id AND users.id= $1", teacher.ID).Scan(&teacher.FirstName, &teacher.Good,
		&teacher.Improve, &teacher.Ahtung, &teacher.Times)
	if err != nil {
		return nil, err
	}

	return &teacher, nil
}

func (p *Postgres) UpdateTeacher(teacher *types.Teacher) error {

	_, err := p.db.Exec("UPDATE users SET first_name = $2, email = $3, updated_at=NOW() WHERE id = $1",
		teacher.ID, teacher.FirstName, teacher.Email)
	if err != nil {
		return errors.Wrap(err, "err with users bd")
	}

	return nil
}

func (p *Postgres) DeleteTeacher(idTeacher int) error {
	//todo как понял для удаления нельзя использовать один запрос. Можно прописать одной строкой через точку с запятой
	//todo но тогда не выловим ошибку на каком именно этапе произошел ерор
	tx, err := p.db.Begin()
	if err != nil {
		return errors.Wrap(err, "err with Begin")
	}
	_, err = tx.Exec("DELETE FROM users WHERE id = $1", idTeacher)
	if err != nil {
		return errors.Wrap(err, "err with delete teacher from users")
	}

	_, err = tx.Exec("DELETE FROM teacher_info WHERE id = $1", idTeacher)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "err with teacher_info bd")
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "err with Commit")
	}
	return nil
}

func (p *Postgres) GetUserByEmail(email string) (*types.User, error) {

	user := types.User{Email: email}
	err := p.db.QueryRow("SELECT id, pass, first_name, user_role FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Password, &user.FirstName, &user.UserRole)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *Postgres) GetTeacherByEmail(email string) (*types.Teacher, error) {

	teacher := types.Teacher{Email: email}
	err := p.db.QueryRow("SELECT id, pass, first_name, user_role FROM users WHERE email = $1", email).
		Scan(&teacher.ID, &teacher.Password, &teacher.FirstName, &teacher.UserRole)
	if err != nil {
		return nil, err
	}

	return &teacher, nil
}

func (p *Postgres) GetUserByID(id int) (*types.User, error) {

	user := types.User{ID: id}
	err := p.db.QueryRow("SELECT email, pass, first_name, user_role FROM users WHERE id = $1", id).
		Scan(&user.Email, &user.Password, &user.FirstName, &user.UserRole)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *Postgres) UpdatePassword(ch *types.ChangePassword) error {

	_, err := p.db.Exec("UPDATE users SET pass = $1, updated_at = NOW() WHERE id = $2", ch.NewPassword, ch.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) AddCodeForRecoveryPass(email, code string) error {

	if _, err := p.db.Exec("INSERT INTO recovery_pass (email, code) VALUES ($1, $2)", email, code); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CheckRecoveryPass(ch *types.RecoveryPasswordEmailAndCode) error {
	//todo Пределать ебала какая то
	if err := p.db.QueryRow("SELECT email, code FROM recovery_pass WHERE email = $1 AND code = $2",
		ch.Email, ch.Code).Scan(&ch.Email, &ch.Code); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteRecoveryPass(email string) error {

	if _, err := p.db.Exec("DELETE FROM recovery_pass WHERE email = $1", email); err != nil {
		return err
	}

	return nil
}

func (p Postgres) CheckDBRecoveryPassword(email string) (bool, error) {
	//todo Пределать ебала какая то
	if err := p.db.QueryRow("SELECT email FROM recovery_pass WHERE email = $1", email).Scan(&email); err != nil {
		return false, err
	}

	return true, nil
}

func (p *Postgres) GetAllStudentsForAdmin() ([]types.UserStat, error) {

	users := make([]types.UserStat, 0)

	rows, err := p.db.Query("SELECT email, first_name, user_role FROM users WHERE user_role = 'student'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	user := types.UserStat{}
	for rows.Next() {
		if err = rows.Scan(&user.Email, &user.FirstName, &user.UserRole); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetAllTeachersForAdmin() ([]types.UserStat, error) {

	users := make([]types.UserStat, 0)

	rows, err := p.db.Query("SELECT email, first_name, user_role FROM users WHERE user_role = 'teacher'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	user := types.UserStat{}
	for rows.Next() {
		if err = rows.Scan(&user.Email, &user.FirstName, &user.UserRole); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetAllAdminsForAdmin() ([]types.UserStat, error) {

	users := make([]types.UserStat, 0)

	rows, err := p.db.Query("SELECT email, first_name, user_role FROM users WHERE user_role = 'admin'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	user := types.UserStat{}
	for rows.Next() {
		if err = rows.Scan(&user.Email, &user.FirstName, &user.UserRole); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetAllUsersForAdmin() ([]types.UserStat, error) {

	users := make([]types.UserStat, 0)

	rows, err := p.db.Query("SELECT email, first_name, user_role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	user := types.UserStat{}
	for rows.Next() {
		if err = rows.Scan(&user.Email, &user.FirstName, &user.UserRole); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (p *Postgres) GetAllTeachersInfoForAdmin() ([]types.TeacherFullInfo, error) {

	teachers := make([]types.TeacherFullInfo, 0)

	rows, err := p.db.Query("SELECT users.id, users.first_name, teacher_info.good, teacher_info.improve, " +
		"teacher_info.ahtung, users.times_seconds FROM users, teacher_info WHERE users.id = teacher_info.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teacher := types.TeacherFullInfo{}
	for rows.Next() {
		if err = rows.Scan(&teacher.ID, &teacher.FirstName, &teacher.Good, &teacher.Improve, &teacher.Ahtung,
			&teacher.Times); err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

func (p *Postgres) GetAllCourse() ([]types.Course, error) {

	courses := make([]types.Course, 0)
	rows, err := p.db.Query("SELECT id, name, cost FROM courses")
	if err != nil {
		return nil, errors.Wrap(err, "err with query")
	}
	defer rows.Close()
	course := types.Course{}
	for rows.Next() {
		if err = rows.Scan(&course.ID, &course.Name, &course.Cost); err != nil {
			return nil, errors.Wrap(err, "err with scan")
		}
		courses = append(courses, course)
	}

	return courses, nil
}

func (p *Postgres) GetAllCoursesInfoForAdmin() ([]types.CourseInfoForAdmin, error) {

	courses := make([]types.CourseInfoForAdmin, 0)
	rows, err := p.db.Query("SELECT id, name, cost, users, dz, sale, total FROM courses")
	if err != nil {
		return nil, errors.Wrap(err, "err with query")
	}
	defer rows.Close()
	course := types.CourseInfoForAdmin{}
	for rows.Next() {
		if err = rows.Scan(&course.ID, &course.Name, &course.Cost, &course.Users, &course.Dz,
			&course.Sale, &course.Total); err != nil {
			return nil, errors.Wrap(err, "err with scan")
		}
		courses = append(courses, course)
	}

	return courses, nil
}

func (p *Postgres) GetAllSectionInCourses(idCourse int) ([]types.Section, error) {

	sections := make([]types.Section, 0)
	rows, err := p.db.Query("SELECT id, course_id, name FROM sections WHERE course_id = $1 ", idCourse)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	section := types.Section{}
	for rows.Next() {
		if err = rows.Scan(&section.ID, &section.CourseID, &section.Name); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func (p *Postgres) GetAllLevelsInSection(idCourse, idSection int) ([]types.Level, error) {

	levels := make([]types.Level, 0)
	rows, err := p.db.Query("SELECT course_id, section_id, level_id, name "+
		"FROM levels WHERE course_id = $1 AND section_id = $2",
		idCourse, idSection)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	level := types.Level{}
	for rows.Next() {
		if err = rows.Scan(&level.CourseID, &level.SectionID, &level.ID,
			&level.Name); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}

		levels = append(levels, level)
	}

	return levels, nil
}

func (p *Postgres) GetAllLessonsInLevel(idCourse, idSection, idLevel int) ([]types.Lesson, error) {

	lessons := make([]types.Lesson, 0)
	rows, err := p.db.Query("SELECT course_id, section_id, level_id, lesson_id, name, description, thesis, task "+
		"FROM lessons WHERE course_id = $1 AND section_id = $2 AND level_id = $3",
		idCourse, idSection, idLevel)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	lesson := types.Lesson{}
	for rows.Next() {
		if err = rows.Scan(&lesson.CourseID, &lesson.SectionID, &lesson.LevelID, &lesson.ID,
			&lesson.Name, &lesson.Description, pq.Array(&lesson.Thesis), &lesson.Task); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}

		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (p *Postgres) CreateCourse(course *types.Course) (*types.OnlyID, error) {

	id := types.OnlyID{}
	if err := p.db.QueryRow("INSERT INTO courses (name, cost, total_price_for_user) VALUES ($1, $2, $3) RETURNING id",
		course.Name, course.Cost, course.TotalPrice).Scan(&id.ID); err != nil {
		return &id, errors.Wrap(err, "err with Exec")
	}

	return &id, nil
}

func (p *Postgres) CreateSection(section *types.Section) (*types.OnlyID, error) {

	id := types.OnlyID{}
	if err := p.db.QueryRow("INSERT INTO sections (course_id, name) VALUES ($1, $2) RETURNING id",
		section.CourseID, section.Name).Scan(&id.ID); err != nil {
		return &id, errors.Wrap(err, "err with Exec")
	}

	return &id, nil
}

func (p *Postgres) CreateLevel(level *types.Level) (*types.OnlyID, error) {

	id := types.OnlyID{}
	if err := p.db.QueryRow("INSERT INTO levels (course_id, section_id, name) VALUES ($1, $2, $3) "+
		"RETURNING level_id", level.CourseID, level.SectionID, level.Name).Scan(&id.ID); err != nil {
		return &id, errors.Wrap(err, "err with Exec")
	}

	return &id, nil
}

func (p *Postgres) CreateLesson(lesson *types.Lesson) (*types.OnlyID, error) {

	id := types.OnlyID{}
	err := p.db.QueryRow("INSERT INTO lessons (course_id, section_id, level_id, name, "+
		"description, thesis, task) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING lesson_id",
		lesson.CourseID, lesson.SectionID, lesson.LevelID, lesson.Name,
		lesson.Description, pq.Array(lesson.Thesis), lesson.Task).Scan(&id.ID)
	if err != nil {
		return &id, errors.Wrap(err, "err with QueryRow added info")
	}

	return &id, nil
}

func (p *Postgres) GetLessonCarousel(idLevel int) (*types.LessonCarousel, error) {

	carousel := types.LessonCarousel{}
	arr := pq.Int64Array{}
	err := p.db.QueryRow("SELECT course_id, section_id, level_id, lesson_array FROM lesson_carousel "+
		"WHERE level_id = $1", idLevel).Scan(&carousel.CourseID, &carousel.SectionID, &carousel.LevelID, &arr)
	if err != nil {
		return nil, err
	}

	carousel.LessonArray = arr

	return &carousel, nil
}

func (p *Postgres) AddCarousel(idCourse, idSection, idLevel int, idLesson []int) error {

	if _, err := p.db.Exec("INSERT INTO lesson_carousel (course_id, section_id, level_id, lesson_array) "+
		"VALUES ($1, $2, $3, $4)", idCourse, idSection, idLevel, pq.Array(idLesson)); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) UpdateCarousel(carousel *types.LessonCarousel) error {

	_, err := p.db.Exec("UPDATE lesson_carousel SET lesson_array = $1 WHERE level_id = $2",
		pq.Array(carousel.LessonArray), carousel.LevelID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteCarousel(levelID int) error {

	_, err := p.db.Exec("DELETE FROM lesson_carousel WHERE level_id = $1", levelID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetCourse(idCourse int) (*types.Course, error) {

	course := types.Course{ID: idCourse}
	err := p.db.QueryRow("SELECT name, cost, sale, total_price_for_user FROM courses WHERE id = $1", idCourse).
		Scan(&course.Name, &course.Cost, &course.Sale, &course.TotalPrice)
	if err != nil {
		return nil, err
	}

	return &course, nil
}

func (p *Postgres) GetSection(idSection int) (*types.Section, error) {

	section := types.Section{ID: idSection}
	err := p.db.QueryRow("SELECT course_id, name FROM sections WHERE id = $1", idSection).
		Scan(&section.CourseID, &section.Name)
	if err != nil {
		return nil, err
	}

	return &section, nil
}

func (p *Postgres) GetLevel(idLevel int) (*types.Level, error) {

	level := types.Level{ID: idLevel}
	err := p.db.QueryRow("SELECT course_id, section_id, name FROM levels WHERE level_id = $1", idLevel).
		Scan(&level.CourseID, &level.SectionID, &level.Name)
	if err != nil {
		return nil, err
	}

	return &level, nil
}

func (p *Postgres) GetLesson(idLesson int) (*types.Lesson, error) {

	lesson := types.Lesson{ID: idLesson}
	err := p.db.QueryRow("SELECT course_id, section_id, level_id, name, description, thesis, task, "+
		"status_free FROM lessons WHERE lesson_id = $1", idLesson).
		Scan(&lesson.CourseID, &lesson.SectionID, &lesson.LevelID, &lesson.Name, &lesson.Description,
			pq.Array(&lesson.Thesis), &lesson.Task, &lesson.Status)
	if err != nil {
		return nil, err
	}

	return &lesson, nil
}

func (p *Postgres) UpdateCourse(course *types.Course) error {

	_, err := p.db.Exec("UPDATE courses SET name = $2, cost = $3, sale = $4, total_price_for_user = $5, "+
		"updated_at = NOW() WHERE id = $1", course.ID, course.Name, course.Cost, course.Sale, course.TotalPrice)
	if err != nil {
		return errors.Wrap(err, "err with Exec")
	}

	return nil
}

func (p *Postgres) UpdateSection(section *types.Section) error {

	if err := p.db.QueryRow("SELECT FROM sections WHERE id = $1", section.ID).Scan(); err != nil {
		return errors.Wrap(err, "err with QueryRow")
	}

	_, err := p.db.Exec("UPDATE sections SET name = $2, updated_at=now() WHERE id = $1", section.ID, section.Name)
	if err != nil {
		return errors.Wrap(err, "err with Exec")
	}

	return nil
}

func (p *Postgres) UpdateLevel(level *types.Level) error {

	if err := p.db.QueryRow("SELECT FROM levels WHERE level_id = $1", level.ID).Scan(); err != nil {
		return errors.Wrap(err, "err with QueryRow")
	}

	_, err := p.db.Exec("UPDATE levels SET name = $2, updated_at=now() WHERE level_id = $1", level.ID, level.Name)
	if err != nil {
		return errors.Wrap(err, "err with Exec")
	}

	return nil
}

func (p *Postgres) UpdateLesson(lesson *types.Lesson) error {

	if err := p.db.QueryRow("SELECT FROM lessons WHERE lesson_id = $1", lesson.ID).Scan(); err != nil {
		return errors.Wrap(err, "err with QueryRow")
	}

	_, err := p.db.Exec("UPDATE lessons SET name = $2, description = $3, thesis = $4, task = $5, "+
		"updated_at = now() WHERE lesson_id = $1",
		lesson.ID, lesson.Name, lesson.Description, pq.Array(lesson.Thesis), lesson.Task)
	if err != nil {
		return errors.Wrap(err, "err with Exec")
	}

	return nil
}

func (p *Postgres) DeleteCourse(idCourse int) error {

	_, err := p.db.Exec("DELETE FROM courses WHERE id = $1", idCourse)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteSection(idCourse, idSection int) error {

	_, err := p.db.Exec("DELETE FROM sections WHERE course_id = $1 AND id = $2", idCourse, idSection)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) DeleteSectionAndTeachersBDBySectionID(idSection int) error {

	_, err := p.db.Exec("DELETE FROM section_and_teacher WHERE section_id = $1", idSection)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteSectionAndTeachersBDByTeacherID(idTeacher int) error {

	_, err := p.db.Exec("DELETE FROM section_and_teacher WHERE teacher_id = $1", idTeacher)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteLevel(idCourse, idSection, idLevel int) error {

	_, err := p.db.Exec("DELETE FROM levels WHERE course_id = $1 AND section_id = $2 AND level_id = $3",
		idCourse, idSection, idLevel)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteLesson(idCourse, idSection, idLevel, idLesson int) error {

	_, err := p.db.Exec("DELETE FROM lessons WHERE "+
		"course_id = $1 AND section_id = $2 AND level_id = $3 AND lesson_id =$4",
		idCourse, idSection, idLevel, idLesson)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetChatsIDByLessonID(lessonID int) ([]int, error) {
	chats := make([]int, 0)
	rows, err := p.db.Query("SELECT chat_id FROM chat WHERE lesson_id = $1", lessonID)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	var chatID int
	for rows.Next() {
		if err = rows.Scan(&chatID); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		chats = append(chats, chatID)
	}

	return chats, nil
}

func (p *Postgres) DeleteChat(chatID int) error {

	_, err := p.db.Exec("DELETE FROM chat WHERE chat_id = $1", chatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) DeleteMessageByChatID(chatID int) error {

	_, err := p.db.Exec("DELETE FROM messages WHERE chat_id = $1", chatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) RecordTime(rec *types.RecordTime) error {
	_, err := p.db.Exec("INSERT INTO request_log (user_id, request_url) VALUES ($1, $2)",
		rec.UserID, rec.RequestURL)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetChatID(chat *types.ChatData) (int, error) {
	var id int

	err := p.db.QueryRow("SELECT chat_id FROM chat WHERE course_id = $1 AND section_id = $2 "+
		"AND level_id = $3 AND lesson_id = $4 AND student_id = $5",
		chat.CourseID, chat.SectionID, chat.LevelID, chat.LessonID, chat.StudentID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) GetMessage(chatID int) ([]types.Message, error) {

	messages := []types.Message{}

	rows, err := p.db.Query("SELECT message_id, text, role, time_mes, first_name FROM messages WHERE chat_id = $1 ", chatID)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	message := types.Message{}
	for rows.Next() {
		if err := rows.Scan(&message.MessageID, &message.Text, &message.Role, &message.TimeMes, &message.FirstName); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (p *Postgres) MakeChat(chat *types.ChatData) (int, error) {

	var id int

	err := p.db.QueryRow("INSERT INTO chat (course_id, section_id, level_id, lesson_id, student_id) "+
		"VALUES ($1, $2, $3, $4, $5) RETURNING chat_id", chat.CourseID, chat.SectionID, chat.LevelID, chat.LessonID,
		chat.StudentID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) SendMessageChat(chatID int, mes *types.MessageBody) error {
	_, err := p.db.Exec("INSERT INTO messages (chat_id, text, role, first_name) VALUES ($1, $2, $3, $4)",
		chatID, mes.Text, mes.Role, mes.FirstName)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetAllSectionsIDByTeacherID(teacherID int) ([]int, error) {

	sectionsID := make([]int, 0)

	rows, err := p.db.Query("SELECT section_id FROM section_and_teacher WHERE teacher_id = $1", teacherID)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	defer rows.Close()
	var sectionID int
	for rows.Next() {
		if err := rows.Scan(&sectionID); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		sectionsID = append(sectionsID, sectionID)
	}

	return sectionsID, nil
}

func (p *Postgres) GetChatsIDBySectionID(sectionID int) ([]int, error) {

	chatsID := make([]int, 0)

	rows, err := p.db.Query("SELECT chat_id FROM chat WHERE section_id = $1", sectionID)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	var chatID int
	for rows.Next() {
		if err := rows.Scan(&chatID); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		chatsID = append(chatsID, chatID)
	}

	return chatsID, nil
}

func (p *Postgres) GetAllChatsIDByStudentID(studentID int) ([]int, error) {

	chatsID := make([]int, 0)

	rows, err := p.db.Query("SELECT chat_id FROM chat WHERE student_id = $1", studentID)
	if err != nil {
		return nil, errors.Wrap(err, "err with Query")
	}
	defer rows.Close()
	var chatID int
	for rows.Next() {
		if err := rows.Scan(&chatID); err != nil {
			return nil, errors.Wrap(err, "err with Scan")
		}
		chatsID = append(chatsID, chatID)
	}

	return chatsID, nil
}

func (p *Postgres) GetStudentIDByChatID(chatID int) (int, error) {

	var studentID int
	err := p.db.QueryRow("SELECT student_id FROM chat WHERE chat_id = $1 ", chatID).Scan(&studentID)
	if err != nil {
		return 0, err
	}

	return studentID, nil
}

func (p *Postgres) GetUserNameByUserID(userID int) (string, error) {

	var UserName string
	err := p.db.QueryRow("SELECT first_name FROM users WHERE id = $1 ", userID).Scan(&UserName)
	if err != nil {
		return "", err
	}

	return UserName, nil
}

func (p *Postgres) GetLessonIDByChatID(chatID int) (int, error) {

	var lessonID int
	err := p.db.QueryRow("SELECT lesson_id FROM chat WHERE chat_id = $1 ", chatID).Scan(&lessonID)
	if err != nil {
		return 0, err
	}

	return lessonID, nil
}

func (p *Postgres) GetLessonNameByLessonID(LessonID int) (string, error) {

	var lessonName string
	err := p.db.QueryRow("SELECT name FROM lessons WHERE lesson_id = $1 ", LessonID).Scan(&lessonName)
	if err != nil {
		return "", err
	}

	return lessonName, nil
}

func (p *Postgres) GetSectionNameBySectionsID(sectionID int) (string, error) {

	var name string
	err := p.db.QueryRow("SELECT name FROM sections WHERE id = $1 ", sectionID).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func (p *Postgres) GetTimeByChatID(chatID int) (string, error) {

	var time string
	err := p.db.QueryRow("SELECT MAX(time_mes) FROM messages WHERE chat_id = $1 ", chatID).Scan(&time)
	if err != nil {
		return "", err
	}

	return time, nil
}

func (p *Postgres) GetAhtungByChatID(chatID int) (bool, error) {

	var ahtung bool
	err := p.db.QueryRow("SELECT ahtung FROM chat WHERE chat_id = $1 ", chatID).Scan(&ahtung)
	if err != nil {
		return false, err
	}

	return ahtung, nil
}

func (p *Postgres) GetSectionIDByChatID(chatID int) (int, error) {

	var sectionID int
	err := p.db.QueryRow("SELECT section_id FROM chat WHERE chat_id = $1 ", chatID).Scan(&sectionID)
	if err != nil {
		return 0, err
	}

	return sectionID, nil
}

func (p *Postgres) GetLastTeacherMessageByChatID(chatID int) (string, error) {

	var lastMessage string
	err := p.db.QueryRow("SELECT text FROM messages WHERE message_id = "+
		"(SELECT MAX(message_id) FROM messages WHERE "+
		"(SELECT MAX(message_id) FROM messages WHERE chat_id = $1) = "+
		"(SELECT MAX(message_id) FROM messages WHERE chat_id = $1 AND role = 'teacher'))",
		chatID).Scan(&lastMessage)
	if err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
		return "", nil
	}
	return lastMessage, nil
}

func (p *Postgres) GetChatDataByChatID(chatID int) (*types.ChatData, error) {

	chatData := types.ChatData{ChatID: chatID}
	err := p.db.QueryRow("SELECT course_id, section_id, level_id, lesson_id, student_id, rating FROM chat "+
		"WHERE chat_id = $1", chatID).Scan(&chatData.CourseID, &chatData.SectionID, &chatData.LevelID,
		&chatData.LessonID, &chatData.StudentID, &chatData.Rating)
	if err != nil {
		return nil, err
	}

	return &chatData, nil
}

func (p *Postgres) CheckURLByC(courseID int) error {

	err := p.db.QueryRow("SELECT id FROM courses WHERE id = $1", courseID).Scan(&courseID)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) CheckURLByCS(courseID, sectionID int) error {

	err := p.db.QueryRow("SELECT course_id, id FROM sections WHERE course_id = $1 AND id = $2",
		courseID, sectionID).Scan(&courseID, &sectionID)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) CheckURLByCSL(courseID, sectionID, levelID int) error {

	err := p.db.QueryRow("SELECT course_id, section_id, level_id FROM levels WHERE "+
		"course_id = $1 AND section_id = $2 AND level_id = $3", courseID, sectionID, levelID).
		Scan(&courseID, &sectionID, &levelID)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) CheckURLByCSLL(courseID, sectionID, levelID, lessonID int) error {

	err := p.db.QueryRow("SELECT course_id, section_id, level_id, lesson_id FROM lessons WHERE "+
		"course_id = $1 AND section_id = $2 AND level_id = $3 AND lesson_id = $4",
		courseID, sectionID, levelID, lessonID).Scan(&courseID, &sectionID, &levelID, &lessonID)
	if err != nil {
		return err
	}

	return nil
}
func (p *Postgres) CheckURLByCSLLC(courseID, sectionID, levelID, lessonID, chatID int) error {

	err := p.db.QueryRow("SELECT course_id, section_id, level_id, lesson_id, chat_id FROM chat WHERE "+
		"course_id = $1 AND section_id = $2 AND level_id = $3 AND lesson_id = $4 AND chat_id = $5",
		courseID, sectionID, levelID, lessonID, chatID).Scan(&courseID, &sectionID, &levelID, &lessonID, &chatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) ChangeAhtung(ch *types.Ahtung) error {

	_, err := p.db.Exec("UPDATE chat SET ahtung_teacher = $1, ahtung = $2 WHERE chat_id = $3",
		ch.TeacherID, ch.Ahtung, ch.ChatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) IncrementAhtung(teacherID int) error {

	_, err := p.db.Exec("UPDATE teacher_info SET ahtung = ahtung + 1 WHERE id = $1", teacherID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) IncrementGood(teacherID int) error {

	_, err := p.db.Exec("UPDATE teacher_info SET good = good + 1 WHERE id = $1", teacherID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) IncrementImprove(teacherID int) error {

	_, err := p.db.Exec("UPDATE teacher_info SET improve = improve + 1 WHERE id = $1", teacherID)
	if err != nil {
		return err
	}

	return nil
}

// Надо получить айди курса
func (p *Postgres) IncrementHomeWork(courseID int) error {

	_, err := p.db.Exec("UPDATE courses SET dz = dz + 1 WHERE id = $1", courseID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) ChangeRating(ch *types.Rating) error {

	_, err := p.db.Exec("UPDATE chat SET rating_teacher = $1, rating = $2 WHERE chat_id = $3",
		ch.TeacherID, ch.Rating, ch.ChatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetRatingByChatID(chatID int) (string, error) {

	var rating string
	if err := p.db.QueryRow("SELECT rating FROM chat WHERE chat_id = $1", chatID).Scan(&rating); err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return rating, err
	}

	return rating, nil
}

func (p *Postgres) GetLastMessageByChatID(chatID int) (string, error) {

	var lastMessage string
	err := p.db.QueryRow("SELECT text FROM messages WHERE message_id = "+
		"(SELECT MAX(message_id) FROM messages WHERE chat_id = $1)", chatID).Scan(&lastMessage)
	if err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
		return "", nil
	}
	return lastMessage, nil
}

func (p *Postgres) OffsetTeacherMessages(chatID int) error {

	_, err := p.db.Exec("UPDATE messages SET not_read = false WHERE chat_id = $1 AND role = 'teacher'", chatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetNotViewMessageForStudentByChatID(chatID int) (int, error) {

	var notViewMes int

	err := p.db.QueryRow("SELECT COUNT (role) FROM messages WHERE role = 'teacher' AND not_read = true "+
		"AND chat_id = $1", chatID).Scan(&notViewMes)
	if err != nil {
		return 0, err
	}

	return notViewMes, err
}

func (p *Postgres) OffsetStudentMessages(chatID int) error {

	_, err := p.db.Exec("UPDATE messages SET not_read = false WHERE chat_id = $1 AND role = 'student'", chatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetNotViewMessageForTeacherByChatID(chatID int) (int, error) {

	var notViewMes int

	err := p.db.QueryRow("SELECT COUNT (role) FROM messages WHERE role = 'student' AND not_read = true "+
		"AND chat_id = $1", chatID).Scan(&notViewMes)
	if err != nil {
		return 0, err
	}

	return notViewMes, err
}

func (p *Postgres) GetCourseIDByChatID(chatID int) (int, error) {

	var courseID int
	if err := p.db.QueryRow("SELECT course_id FROM chat WHERE chat_id = $1", chatID).Scan(&courseID); err != nil {
		return 0, err
	}

	return courseID, nil
}

func (p *Postgres) LastSession(userID int) (string, error) {

	var date string
	if err := p.db.QueryRow("SELECT MAX(created_at) FROM request_log WHERE user_id = $1", userID).Scan(&date); err != nil {
		return "", err
	}

	return date, nil
}

func (p *Postgres) AddTime(seconds, userID int) error {

	if _, err := p.db.Exec("UPDATE users SET times_seconds = times_seconds + $1 WHERE id = $2", seconds, userID); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) CheckUserInDBUsers(userID int) error {

	if err := p.db.QueryRow("SELECT FROM users WHERE id = $1", userID).Scan(); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) FindLastStudentMessageNotAnswer(chatID int) (string, error) {

	var time string
	if err := p.db.QueryRow("SELECT time_mes FROM messages WHERE chat_id = $1 AND "+
		"message_id = (SELECT MAX(message_id) FROM messages "+
		"WHERE "+
		"("+
		"(SELECT MAX(message_id) FROM messages WHERE chat_id = $1) "+
		"= "+
		"(SELECT MAX(message_id) FROM messages WHERE chat_id = $1 AND role = 'student'))"+
		")", chatID).Scan(&time); err != nil {
		return "", err
	}

	return time, nil
}

func (p *Postgres) AddTimeAnswer(teacherID, seconds int) error {

	if _, err := p.db.Exec("UPDATE teacher_info SET answer_time_sec = answer_time_sec + $1, "+
		"answer_count = answer_count + 1 WHERE id = $2", seconds, teacherID); err != nil {
		return err
	}

	return nil
}

func (p *Postgres) GetAverageTimeByTeacherID(teacherID int) (*types.AverageTime, error) {

	averageTime := &types.AverageTime{}
	if err := p.db.QueryRow("SELECT answer_time_sec, answer_count FROM teacher_info WHERE id = $1", teacherID).
		Scan(&averageTime.TotalTimeForAnswer, &averageTime.CountAnswer); err != nil {
		return nil, err
	}

	return averageTime, nil
}
