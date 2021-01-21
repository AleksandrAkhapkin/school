package server

import (
	"github.com/gorilla/mux"
	"github.com/tarasova-school/internal/tarasova-school/server/handlers"
	"net/http"
)

func NewRouter(h *handlers.Handlers) *mux.Router {
	//todo надо проверить логичность всех путей
	router := mux.NewRouter().StrictSlash(true)
	router.Use(h.RecoverPanic)
	router.Use(h.RecordRequest)
	adminRouter := router.PathPrefix("").Subrouter()
	teacherRouter := router.PathPrefix("").Subrouter()
	studentRouter := router.PathPrefix("").Subrouter()
	adminAndTeacherRouter := router.PathPrefix("").Subrouter()
	adminRouter.Use(h.CheckRoleAdmin)
	teacherRouter.Use(h.CheckRoleTeacher)
	studentRouter.Use(h.CheckRoleStudent)
	adminAndTeacherRouter.Use(h.CheckRoleAdminAndTeacher)
	adminRouter.Use(h.CheckUserInDBUsers)
	teacherRouter.Use(h.CheckUserInDBUsers)
	studentRouter.Use(h.CheckUserInDBUsers)
	adminAndTeacherRouter.Use(h.CheckUserInDBUsers)

	adminRouter.Methods(http.MethodPost).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}/upload").HandlerFunc(h.UploadVideo)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}/video").HandlerFunc(h.GetVideo)

	router.Methods(http.MethodGet).Path("/ping").HandlerFunc(h.Ping)
	router.Methods(http.MethodPost).Path("/users/auth").HandlerFunc(h.Authorize)
	router.Methods(http.MethodPost).Path("/users/register").HandlerFunc(h.RegisterUser)
	router.Methods(http.MethodPost).Path("/password/change").HandlerFunc(h.ChangePassword)
	router.Methods(http.MethodPost).Path("/password/recovery").HandlerFunc(h.RecoveryPassword)
	router.Methods(http.MethodPost).Path("/password/recovery/check").HandlerFunc(h.CheckValidRecoveryPassword)
	router.Methods(http.MethodPost).Path("/password/recovery/new").HandlerFunc(h.NewRecoveryPassword)

	router.Methods(http.MethodGet).Path("/vk/callback").HandlerFunc(h.VKCallback)

	adminRouter.Methods(http.MethodPost).Path("/users/register/teacher").HandlerFunc(h.RegisterTeacher)
	adminRouter.Methods(http.MethodGet).Path("/users/teacher/{idTeacher:[0-9]+}").HandlerFunc(h.GetTeacher)
	adminRouter.Methods(http.MethodPut).Path("/users/teacher/{idTeacher:[0-9]+}").HandlerFunc(h.UpdateTeacher)
	adminRouter.Methods(http.MethodDelete).Path("/users/teacher/{idTeacher:[0-9]+}").HandlerFunc(h.DeleteTeacher)

	adminRouter.Methods(http.MethodGet).Path("/admin/courses/all").HandlerFunc(h.GetAllCoursesInfoForAdmin)
	adminRouter.Methods(http.MethodGet).Path("/admin/teachers/all").HandlerFunc(h.GetAllTeachersInfoForAdmin)

	adminAndTeacherRouter.Methods(http.MethodGet).Path("/users").HandlerFunc(h.GetUsers)
	//todo дописать в сваггер GetStudentByQueryID, проверить поля в постгрессе
	router.Methods(http.MethodGet).Path("/users/{idStudent:[0-9]+}").HandlerFunc(h.GetStudentByQueryID)
	router.Methods(http.MethodGet).Path("/courses/all").HandlerFunc(h.GetAllCourses)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/all").HandlerFunc(h.GetAllSectionsInCourse)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/all").HandlerFunc(h.GetAllLevelsInSection)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/all").HandlerFunc(h.GetAllLessonsInLevel)

	adminRouter.Methods(http.MethodPost).Path("/courses").HandlerFunc(h.AddCourse)
	adminRouter.Methods(http.MethodPost).Path("/courses/{idCourse:[0-9]+}/sections").HandlerFunc(h.AddSection)
	adminRouter.Methods(http.MethodPost).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels").HandlerFunc(h.AddLevel)
	adminRouter.Methods(http.MethodPost).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons").HandlerFunc(h.AddLesson)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}").HandlerFunc(h.GetCourse)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}").HandlerFunc(h.GetSection)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}").HandlerFunc(h.GetLevel)
	router.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}").HandlerFunc(h.GetLesson)
	adminRouter.Methods(http.MethodPut).Path("/courses/{idCourse:[0-9]+}").HandlerFunc(h.UpdateCourse)
	adminRouter.Methods(http.MethodPut).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}").HandlerFunc(h.UpdateSection)
	adminRouter.Methods(http.MethodPut).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}").HandlerFunc(h.UpdateLevel)
	adminRouter.Methods(http.MethodPut).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}").HandlerFunc(h.UpdateLesson)
	adminRouter.Methods(http.MethodDelete).Path("/courses/{idCourse:[0-9]+}").HandlerFunc(h.DeleteCourse)
	adminRouter.Methods(http.MethodDelete).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}").HandlerFunc(h.DeleteSection)
	adminRouter.Methods(http.MethodDelete).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}").HandlerFunc(h.DeleteLevel)
	adminRouter.Methods(http.MethodDelete).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}").HandlerFunc(h.DeleteLesson)

	//получить чат на странице урока
	studentRouter.Methods(http.MethodGet).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}/chat").HandlerFunc(h.GetChatForStudentByLesson)
	//отправка сообщения (начать или продолжить чат на странице урока)
	studentRouter.Methods(http.MethodPost).Path("/courses/{idCourse:[0-9]+}/sections/{idSection:[0-9]+}/levels/{idLevel:[0-9]+}/lessons/{idLesson:[0-9]+}/chat").HandlerFunc(h.SendMessageToChatByLessonForStudent)
	//показать превью чатов которые уже были начаты
	studentRouter.Methods(http.MethodGet).Path("/student/chat/all").HandlerFunc(h.GetAllChatsForStudent)
	//получить чат из превью в личном кабинете
	studentRouter.Methods(http.MethodGet).Path("/student/chat/{idChat:[0-9]+}").HandlerFunc(h.GetChatByProfileStudent)
	//отправить сообщение в уже существующий чат полученный из превью в личном кабинете
	studentRouter.Methods(http.MethodPost).Path("/student/chat/{idChat:[0-9]+}").HandlerFunc(h.SendMessageToChatByProfileStudent)

	//показать чаты в разделе чаты
	teacherRouter.Methods(http.MethodGet).Path("/teacher/chat/all").HandlerFunc(h.GetAllChatsForTeacher)
	//показать чаты в инфо блоке на страничке учителя
	adminRouter.Methods(http.MethodGet).Path("/users/teacher/{idTeacher:[0-9]+}/chats").HandlerFunc(h.GetAllChatsForAdmin)
	//показать чат на страничке учителя
	adminRouter.Methods(http.MethodGet).Path("/users/teacher/{idTeacher:[0-9]+}/chat/{idChat:[0-9]+}").HandlerFunc(h.GetChatForAdmin)
	//получить конкретный чат из превью
	teacherRouter.Methods(http.MethodGet).Path("/teacher/chat/{idChat:[0-9]+}").HandlerFunc(h.GetChatForTeacher)
	//отправить сообщение в конкретный чат из превью
	teacherRouter.Methods(http.MethodPost).Path("/teacher/chat/{idChat:[0-9]+}").HandlerFunc(h.SendMessageToChatForTeacher)

	teacherRouter.Methods(http.MethodPost).Path("/teacher/chat/{idChat:[0-9]+}/ahtung").HandlerFunc(h.Ahtung)
	teacherRouter.Methods(http.MethodPost).Path("/teacher/chat/{idChat:[0-9]+}/rating").HandlerFunc(h.Rating)

	return router
}
