package types

import (
	"mime/multipart"
)

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

const (
	EmailRecoveryTitle = "Востановление пароля для tarasova-school.ru"
	MIME               = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type ChangePassword struct {
	UserID         int    `json:"user_id"`
	OldPassword    string `json:"old_password"`
	NewPassword    string `json:"new_password"`
	RepeatPassword string `json:"repeat_password"`
}

type RecoveryPasswordEmail struct {
	Email string `json:"email"`
}

type RecoveryPasswordEmailAndCode struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type RecoveryPasswordNewPass struct {
	Email          string `json:"email"`
	Code           string `json:"code"`
	NewPassword    string `json:"new_password"`
	RepeatPassword string `json:"repeat_password"`
}

type CheckCode struct {
	Code bool `json:"code"`
}

type AuthorizeVK struct {
	Email     string
	Firstname string
}

type Authorize struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type User struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	ID        int    `json:"id"`
	UserRole  string `json:"role"`
	CreatedAT string `json:"created_at"`
	UpdatedAT string `json:"updated_at"`
}

type Teacher struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	ID        int    `json:"id"`
	UserRole  string `json:"role"`
	CreatedAT string `json:"created_at"`
	UpdatedAT string `json:"updated_at"`
}

type TeacherFullInfo struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	Good        int    `json:"good"`
	Improve     int    `json:"improve"`
	Ahtung      int    `json:"ahtung"`
	Times       int    `json:"times"`
	AverageTime int    `json:"average_time"`
}

type AverageTime struct {
	TotalTimeForAnswer int
	CountAnswer        int
}

type Admin struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	ID        int    `json:"id"`
	UserRole  string `json:"role"`
	CreatedAT string `json:"created_at"`
	UpdatedAT string `json:"updated_at"`
}

type UserStat struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	UserRole  string `json:"role"`
}

type Course struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Cost       int    `json:"cost"`
	Sale       int    `json:"sale"`
	TotalPrice int    `json:"total_price"`
}

type CourseInfoForAdmin struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Cost  string `json:"cost"`
	Users int    `json:"users"`
	Dz    int    `json:"dz"`
	Sale  int    `json:"sale"`
	Total int    `json:"total"`
}

type Section struct {
	ID       int    `json:"id"`
	CourseID int    `json:"course_id"`
	Name     string `json:"name"`
}

type Level struct {
	ID        int    `json:"id"`
	CourseID  int    `json:"course_id"`
	SectionID int    `json:"section_id"`
	Name      string `json:"name"`
}

type Lesson struct {
	ID        int `json:"id"`
	CourseID  int `json:"course_id"`
	SectionID int `json:"section_id"`
	LevelID   int `json:"level_id"`

	Name        string   `json:"name"`
	Description string   `json:"lesson_description"`
	Thesis      []string `json:"lesson_thesis"`
	Task        string   `json:"lesson_task"`

	Status        bool   `json:"status_free"`
	NextLessonID  int    `json:"next_lesson_id"`
	NextLessonURL string `json:"next_lesson_url"`
}

type LessonCarousel struct {
	LessonArray []int64 `json:"id"`
	CourseID    int     `json:"course_id"`
	SectionID   int     `json:"section_id"`
	LevelID     int     `json:"level_id"`
}

type OnlyID struct {
	ID int `json:"id"`
}

type RecordTime struct {
	UserID     int
	RequestURL string
	Time       string
}

type ChatData struct {
	CourseID  int       `json:"course_id"`
	SectionID int       `json:"section_id"`
	LevelID   int       `json:"level_id"`
	LessonID  int       `json:"lesson_id"`
	ChatID    int       `json:"chat_id"`
	StudentID int       `json:"student_id"`
	Rating    string    `json:"rating"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	Role      string `json:"role"`
	TimeMes   string `json:"time_mes"`
	FirstName string `json:"first_name"`
}

type MessageBody struct {
	Text      string `json:"text"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	UserID    int    `json:"user_id"`
}

type ChatsPreviewForStudent struct {
	ChatID         int    `json:"chat_id"`
	SectionsID     int    `json:"sections_id"`
	SectionsName   string `json:"sections_name"`
	LessonID       int    `json:"lesson_id"`
	LessonName     string `json:"lesson_name"`
	LastMessage    string `json:"last_message"`
	NotViewMessage int    `json:"not_view_message"`
}

type ChatsPreviewForTeacher struct {
	ChatID           int    `json:"chat_id"`
	StudentID        int    `json:"student_id"`
	StudentFirstName string `json:"student_first_name"`
	SectionsID       int    `json:"sections_id"`
	SectionsName     string `json:"sections_name"`
	LessonID         int    `json:"lesson_id"`
	LessonName       string `json:"lesson_name"`
	Time             string `json:"time"`
	Ahtung           bool   `json:"ahtung"`
	NotViewMessage   int    `json:"not_view_message"`
}

type ChatsPreviewForAdmin struct {
	ChatID           int    `json:"chat_id"`
	StudentID        int    `json:"student_id"`
	StudentFirstName string `json:"student_first_name"`
	SectionsID       int    `json:"sections_id"`
	SectionsName     string `json:"sections_name"`
	LessonID         int    `json:"lesson_id"`
	LessonName       string `json:"lesson_name"`
	Time             string `json:"time"`
	Rating           string `json:"rating"`
	LastMessage      string `json:"last_message"`
	NotViewMessage   int    `json:"not_view_message"`
}

type Ahtung struct {
	ChatID    int  `json:"chat_id"`
	TeacherID int  `json:"teacher_id"`
	Ahtung    bool `json:"ahtung"`
}

type Rating struct {
	ChatID    int    `json:"chat_id"`
	TeacherID int    `json:"teacher_id"`
	Rating    string `json:"rating"` //improve or good or ""
}

type UploadVideo struct {
	CourseID  int `json:"course_id"`
	SectionID int `json:"section_id"`
	LevelID   int `json:"level_id"`
	LessonID  int `json:"lesson_id"`

	Body multipart.File
}

type GetVideo struct {
	CourseID  int `json:"course_id"`
	SectionID int `json:"section_id"`
	LevelID   int `json:"level_id"`
	LessonID  int `json:"lesson_id"`
}
