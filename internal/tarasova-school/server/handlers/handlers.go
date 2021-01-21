package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tarasova-school/internal/tarasova-school/service"
	"github.com/tarasova-school/internal/types"
	"github.com/tarasova-school/internal/types/config"
	"github.com/tarasova-school/pkg/infrastruct"
	"github.com/tarasova-school/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Handlers struct {
	srv       *service.Service
	soc       *config.SocAuth
	secretKey string
}

func NewHandlers(srv *service.Service, cnf *config.Config) *Handlers {

	return &Handlers{
		srv:       srv,
		secretKey: cnf.SecretKeyJWT,
		soc:       cnf.Soc,
	}
}

func (h *Handlers) Ping(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("pong"))
}

func (h *Handlers) UploadVideo(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	src, _, err := r.FormFile("video")
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with FormFile in UploadVideo"))
		return
	}
	video := types.UploadVideo{}

	video.CourseID = idCourse
	video.SectionID = idSection
	video.LevelID = idLevel
	video.LessonID = idLesson
	video.Body = src

	defer src.Close()

	err = h.srv.UploadVideo(&video)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

}

func (h *Handlers) GetVideo(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	video := types.GetVideo{}

	video.CourseID = idCourse
	video.SectionID = idSection
	video.LevelID = idLevel
	video.LessonID = idLesson

	url, err := h.srv.GetVideoURL(&video)
	videoStream, _ := os.Open(url)

	io.Copy(w, videoStream)
}

func (h *Handlers) Authorize(w http.ResponseWriter, r *http.Request) {

	auth := types.Authorize{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&auth); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	auth.Email = strings.ToLower(auth.Email)

	token, err := h.srv.Authorize(&auth)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, token)
}

func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {

	user := types.User{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	user.Email = strings.ToLower(user.Email)

	if user.Email == "" || user.FirstName == "" || user.Password == "" {
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	token, err := h.srv.RegisterStudent(&user)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, token)
}

func (h *Handlers) RegisterTeacher(w http.ResponseWriter, r *http.Request) {

	teacher := types.Teacher{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	teacher.Email = strings.ToLower(teacher.Email)

	if teacher.Email == "" || teacher.FirstName == "" || teacher.Password == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	id, err := h.srv.RegisterTeacher(&teacher)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, id)
}

func (h *Handlers) GetTeacher(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idTeacher, err := strconv.Atoi(query["idTeacher"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
	}

	teacher, err := h.srv.GetTeacher(idTeacher)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, teacher)
}

func (h *Handlers) UpdateTeacher(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idTeacher, err := strconv.Atoi(query["idTeacher"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
	}

	teacher := types.Teacher{}
	if err := json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	teacher.Email = strings.ToLower(teacher.Email)

	if teacher.FirstName == "" || teacher.Email == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
	}

	teacher.ID = idTeacher

	err = h.srv.UpdateTeacher(&teacher)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) DeleteTeacher(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idTeacher, err := strconv.Atoi(query["idTeacher"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
	}

	err = h.srv.DeleteTeacher(idTeacher)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request) {

	ch := types.ChangePassword{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	if ch.NewPassword != ch.RepeatPassword {
		apiResponseEncoder(w, infrastruct.ErrorPasswordsDoNotMatch)
		return
	}
	if err = h.srv.ChangePassword(&ch); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) RecoveryPassword(w http.ResponseWriter, r *http.Request) {

	ch := types.RecoveryPasswordEmail{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	ch.Email = strings.ToLower(ch.Email)
	ch.Email = strings.TrimSpace(ch.Email)

	if err = h.srv.RecoveryPassword(&ch); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) CheckValidRecoveryPassword(w http.ResponseWriter, r *http.Request) {

	ch := types.RecoveryPasswordEmailAndCode{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	ch.Email = strings.ToLower(ch.Email)
	ch.Email = strings.TrimSpace(ch.Email)
	ch.Code = strings.ToUpper(ch.Code)

	ok, err := h.srv.CheckValidRecoveryPassword(&ch)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	check := &types.CheckCode{Code: ok}

	apiResponseEncoder(w, check)
}

func (h *Handlers) NewRecoveryPassword(w http.ResponseWriter, r *http.Request) {

	ch := types.RecoveryPasswordNewPass{}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&ch); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	ch.Email = strings.ToLower(ch.Email)
	ch.Email = strings.TrimSpace(ch.Email)

	if err = h.srv.NewRecoveryPassword(&ch); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	bb, _ := ioutil.ReadAll(r.Body)
	logger.LogInfo(string(bb))
}

func (h *Handlers) GetStudentByQueryID(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idStudent, err := strconv.Atoi(query["idStudent"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	user, err := h.srv.GetStudentByID(idStudent)
	if err != nil {

	}
	apiResponseEncoder(w, user)
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {

	userRole := r.FormValue("role")
	moreUser, err := h.srv.GetUsers(userRole)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
	apiResponseEncoder(w, moreUser)
}

func (h *Handlers) GetAllTeachersInfoForAdmin(w http.ResponseWriter, r *http.Request) {

	teacherArr, err := h.srv.GetAllTeachersInfoForAdmin()
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
	apiResponseEncoder(w, teacherArr)
}

func (h *Handlers) GetAllCourses(w http.ResponseWriter, _ *http.Request) {

	courseArr, err := h.srv.GetAllCourse()
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, courseArr)
}

func (h *Handlers) GetAllCoursesInfoForAdmin(w http.ResponseWriter, _ *http.Request) {

	courseArr, err := h.srv.GetAllCoursesInfoForAdmin()
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, courseArr)
}

func (h *Handlers) GetAllSectionsInCourse(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	sectionArr, err := h.srv.GetAllSectionsInCourse(idCourse)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, sectionArr)
}

func (h *Handlers) GetAllLevelsInSection(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	levelArr, err := h.srv.GetAllLevelsInSection(idCourse, idSection)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, levelArr)
}

func (h *Handlers) GetAllLessonsInLevel(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lessonArr, err := h.srv.GetAllLessonsInLevel(idCourse, idSection, idLevel)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, lessonArr)
}

func (h *Handlers) AddCourse(w http.ResponseWriter, r *http.Request) {

	course := types.Course{}

	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if course.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	id, err := h.srv.AddCourse(&course)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, id)
}

func (h *Handlers) AddSection(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	section := types.Section{}

	if err = json.NewDecoder(r.Body).Decode(&section); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	section.CourseID = idCourse

	if section.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	idSection, err := h.srv.AddSection(&section)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, idSection)
}

func (h *Handlers) AddLevel(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	level := types.Level{}

	if err = json.NewDecoder(r.Body).Decode(&level); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	level.CourseID = idCourse
	level.SectionID = idSection

	if level.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	idLevel, err := h.srv.AddLevel(&level)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, idLevel)
}

func (h *Handlers) AddLesson(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lesson := types.Lesson{}

	if err = json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lesson.CourseID = idCourse
	lesson.SectionID = idSection
	lesson.LevelID = idLevel

	if lesson.Name == "" || lesson.Description == "" || len(lesson.Thesis) == 0 || lesson.Task == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	idLesson, err := h.srv.AddLesson(&lesson)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, idLesson)
}

func (h *Handlers) GetCourse(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	course, err := h.srv.GetCourse(idCourse)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, course)
}

func (h *Handlers) GetSection(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	section, err := h.srv.GetSection(idCourse, idSection)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, section)
}

func (h *Handlers) GetLevel(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	level, err := h.srv.GetLevel(idCourse, idSection, idLevel)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, level)
}

func (h *Handlers) GetLesson(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lesson, err := h.srv.GetLesson(idCourse, idSection, idLevel, idLesson)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, lesson)
}

func (h *Handlers) UpdateCourse(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	course := types.Course{}
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	course.ID = idCourse

	if course.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.UpdateCourse(&course); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) UpdateSection(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	section := types.Section{}
	if err := json.NewDecoder(r.Body).Decode(&section); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	section.CourseID = idCourse
	section.ID = idSection

	if section.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.UpdateSection(&section); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) UpdateLevel(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	level := types.Level{}
	if err := json.NewDecoder(r.Body).Decode(&level); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	level.CourseID = idCourse
	level.SectionID = idSection
	level.ID = idLevel

	if level.Name == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.UpdateLevel(&level); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) UpdateLesson(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lesson := types.Lesson{}
	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	lesson.CourseID = idCourse
	lesson.SectionID = idSection
	lesson.LevelID = idLevel
	lesson.ID = idLesson

	if lesson.Name == "" || lesson.Description == "" || len(lesson.Thesis) == 0 || lesson.Task == "" {
		logger.LogError(infrastruct.ErrorBadRequest)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.UpdateLesson(&lesson); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) DeleteCourse(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.DeleteCourse(idCourse); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) DeleteSection(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.DeleteSection(idCourse, idSection); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) DeleteLevel(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.DeleteLevel(idCourse, idSection, idLevel); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) DeleteLesson(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	if err := h.srv.DeleteLesson(idCourse, idSection, idLevel, idLesson); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) GetChatForStudentByLesson(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	chat := &types.ChatData{CourseID: idCourse,
		SectionID: idSection,
		LevelID:   idLevel,
		LessonID:  idLesson,
		StudentID: claims.UserID,
	}

	messages, err := h.srv.GetChatByLessonForStudent(chat)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, messages)
}

func (h *Handlers) SendMessageToChatByLessonForStudent(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idCourse, err := strconv.Atoi(query["idCourse"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idSection, err := strconv.Atoi(query["idSection"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLevel, err := strconv.Atoi(query["idLevel"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}
	idLesson, err := strconv.Atoi(query["idLesson"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	chat := &types.ChatData{CourseID: idCourse,
		SectionID: idSection,
		LevelID:   idLevel,
		LessonID:  idLesson,
		StudentID: claims.UserID,
	}
	text := types.MessageBody{}
	if err = json.NewDecoder(r.Body).Decode(&text); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}
	text.Role = claims.Role
	text.UserID = claims.UserID

	if err = h.srv.SendMessageToChatByLessonForStudent(chat, &text); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) GetChatByProfileStudent(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	messages, err := h.srv.GetChatByProfileStudent(idChat, claims)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, messages)
}

func (h *Handlers) SendMessageToChatByProfileStudent(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	text := types.MessageBody{}
	if err = json.NewDecoder(r.Body).Decode(&text); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
	text.Role = claims.Role
	text.UserID = claims.UserID

	if err = h.srv.SendMessageToChatByProfileStudent(idChat, &text); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) GetChatForTeacher(w http.ResponseWriter, r *http.Request) {
	//todo мб объеденить с гет чатом студента если методы в сервисе будут одинаковые

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	messages, err := h.srv.GetChatForTeacher(idChat, claims)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, messages)
}

func (h *Handlers) GetChatForAdmin(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	messages, err := h.srv.GetChatForAdmin(idChat)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, messages)
}

func (h *Handlers) SendMessageToChatForTeacher(w http.ResponseWriter, r *http.Request) {
	//todo мб объединить с сенд месседж студента если в сервисе будут одинаковые методы

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	text := types.MessageBody{}
	if err = json.NewDecoder(r.Body).Decode(&text); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}
	text.Role = claims.Role
	text.UserID = claims.UserID

	if err = h.srv.SendMessageToChatForTeacher(idChat, &text); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h Handlers) GetAllChatsForStudent(w http.ResponseWriter, r *http.Request) {

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	previewChats, err := h.srv.GetAllChatsForStudent(claims.UserID)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, previewChats)
}

func (h Handlers) GetAllChatsForTeacher(w http.ResponseWriter, r *http.Request) {

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	previewChats, err := h.srv.GetAllChatsForTeacher(claims.UserID)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, previewChats)
}

func (h Handlers) GetAllChatsForAdmin(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idTeacher, err := strconv.Atoi(query["idTeacher"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	previewChats, err := h.srv.GetAllChatsForAdmin(idTeacher)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, previewChats)
}

func (h *Handlers) Ahtung(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	ahtung := &types.Ahtung{ChatID: idChat, TeacherID: claims.UserID}

	if err = h.srv.Ahtung(ahtung); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func (h *Handlers) Rating(w http.ResponseWriter, r *http.Request) {

	query := mux.Vars(r)
	idChat, err := strconv.Atoi(query["idChat"])
	if err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorBadRequest)
		return
	}

	claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	rating := &types.Rating{}
	if err = json.NewDecoder(r.Body).Decode(&rating); err != nil {
		logger.LogError(err)
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	rating.TeacherID = claims.UserID
	rating.ChatID = idChat

	if err = h.srv.Rating(rating); err != nil {
		apiErrorEncode(w, err)
		return
	}
}

func apiErrorEncode(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if customError, ok := err.(*infrastruct.CustomError); ok {
		w.WriteHeader(customError.Code)
	}

	result := struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}

	if err = json.NewEncoder(w).Encode(result); err != nil {
		logger.LogError(err)
	}
}

func apiResponseEncoder(w http.ResponseWriter, res interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.LogError(err)
	}
}
