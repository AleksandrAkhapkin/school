package service

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/pkg/errors"
	"github.com/tarasova-school/internal/clients/postgres"
	"github.com/tarasova-school/internal/tarasova-school/service/mail"
	"github.com/tarasova-school/internal/types"
	"github.com/tarasova-school/internal/types/config"
	"github.com/tarasova-school/pkg/infrastruct"
	"github.com/tarasova-school/pkg/logger"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	p                *postgres.Postgres
	secretKey        string
	email            *config.ConfigForSendEmail
	htmlRecoveryPath string
	htmlNewPassPath  string
	videoDir         string
}

func NewService(pg *postgres.Postgres, cnf *config.Config) (*Service, error) {

	return &Service{
		p:                pg,
		secretKey:        cnf.SecretKeyJWT,
		email:            cnf.Email,
		htmlRecoveryPath: cnf.HtmlRecoveryPath,
		htmlNewPassPath:  cnf.HtmlNewPassPath,
		videoDir:         cnf.VideoDir,
	}, nil
}

func (s *Service) AuthorizeVK(auth *types.AuthorizeVK) (*types.Token, error) {
	user, err := s.p.GetUserByEmail(auth.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(err)
			return nil, infrastruct.ErrorInternalServerError
		}

		newPass, err := uuid.GenerateUUID()
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with GenerateUUID"))
			return nil, infrastruct.ErrorInternalServerError
		}
		user := &types.User{
			FirstName: auth.Firstname,
			Email:     auth.Email,
			Password:  newPass,
			UserRole:  types.RoleStudent,
		}
		id, err := s.p.CreateUser(user)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with CreateUser"))
			return nil, infrastruct.ErrorInternalServerError
		}
		user.ID = id

		r := mail.NewRequest([]string{user.Email}, s.email)
		if err = r.Send(s.htmlNewPassPath, newPass); err != nil {
			logger.LogError(errors.Wrap(err, "err with send Email"))
			return nil, infrastruct.ErrorInternalServerError
		}

		token, err := infrastruct.GenerateJWT(user.ID, user.UserRole, s.secretKey)
		if err != nil {
			logger.LogError(err)
			return nil, infrastruct.ErrorInternalServerError
		}

		return &types.Token{Token: token}, nil
	}

	token, err := infrastruct.GenerateJWT(user.ID, user.UserRole, s.secretKey)
	if err != nil {
		logger.LogError(err)
		return nil, infrastruct.ErrorInternalServerError
	}

	return &types.Token{Token: token}, nil
}

func (s *Service) Authorize(auth *types.Authorize) (*types.Token, error) {
	user, err := s.p.GetUserByEmail(strings.TrimSpace(auth.Email))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorPasswordIsIncorrect
		}
		logger.LogError(err)
		return nil, infrastruct.ErrorInternalServerError
	}

	if user.Password != strings.TrimSpace(auth.Password) {
		return nil, infrastruct.ErrorPasswordIsIncorrect
	}

	token, err := infrastruct.GenerateJWT(user.ID, user.UserRole, s.secretKey)
	if err != nil {
		logger.LogError(err)
		return nil, infrastruct.ErrorInternalServerError
	}

	return &types.Token{Token: token}, nil
}

func (s *Service) GetVideoURL(video *types.GetVideo) (string, error) {

	//check correct courseID in URL
	if err := s.p.CheckURLByCSLL(video.CourseID, video.SectionID, video.LevelID, video.LessonID); err != nil {
		return "", infrastruct.ErrorNotFound
	}

	url := fmt.Sprintf("%s/%d", s.videoDir, video.LessonID)

	return url, nil
}

func (s *Service) RegisterStudent(user *types.User) (*types.Token, error) {
	trimSpaceUser(user)
	_, err := s.p.GetUserByEmail(user.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(err)
			return nil, infrastruct.ErrorInternalServerError
		}
	} else {
		return nil, infrastruct.ErrorEmailIsExist
	}

	user.UserRole = types.RoleStudent
	id, err := s.p.CreateUser(user)
	if err != nil {
		logger.LogError(err)
		return nil, infrastruct.ErrorInternalServerError
	}

	user.ID = id

	logger.SendMessage(fmt.Sprintf("Зарегистрировался новый студент: %s, email: %s",
		user.FirstName, user.Email))

	token, err := infrastruct.GenerateJWT(user.ID, user.UserRole, s.secretKey)
	if err != nil {
		logger.LogError(err)
		return nil, infrastruct.ErrorInternalServerError
	}

	return &types.Token{Token: token}, nil
}

func (s *Service) RegisterTeacher(teacher *types.Teacher) (*types.OnlyID, error) {
	trimSpaceTeacher(teacher)
	_, err := s.p.GetUserByEmail(teacher.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, infrastruct.ErrorInternalServerError
		}
	} else {
		return nil, infrastruct.ErrorEmailIsExist
	}

	teacher.UserRole = types.RoleTeacher
	err = s.p.CreateTeacher(teacher)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with CreateTeacher"))
		return nil, infrastruct.ErrorInternalServerError
	}

	teacher, err = s.p.GetTeacherByEmail(teacher.Email)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetTeacherByEmail"))
		return nil, infrastruct.ErrorInternalServerError
	}

	id := &types.OnlyID{ID: teacher.ID}
	logger.SendMessage(fmt.Sprintf("Новый учитель успешно зарегистрирован: %s, email: %s",
		teacher.FirstName, teacher.Email))

	return id, nil
}

func (s *Service) GetTeacher(idTeacher int) (*types.TeacherFullInfo, error) {

	teacher, err := s.p.GetTeacherByID(idTeacher)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(err)
			return nil, infrastruct.ErrorInternalServerError
		}
		return nil, infrastruct.ErrorNotFound
	}

	teacher.Times = teacher.Times / 60 / 60

	averageTime, err := s.p.GetAverageTimeByTeacherID(idTeacher)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with s.p.GetAverageTimeByTeacherID"))
		return nil, infrastruct.ErrorInternalServerError
	}

	if averageTime.CountAnswer == 0 {
		return teacher, nil
	}

	teacher.AverageTime = (averageTime.TotalTimeForAnswer / averageTime.CountAnswer) / 60

	return teacher, nil
}

func (s *Service) UpdateTeacher(teacher *types.Teacher) error {

	//check correct idTeacher
	if _, err := s.GetTeacher(teacher.ID); err != nil {
		return err
	}

	trimSpaceTeacher(teacher)
	if err := s.p.UpdateTeacher(teacher); err != nil {
		logger.LogError(err)
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) DeleteTeacher(idTeacher int) error {

	if err := s.p.DeleteTeacher(idTeacher); err != nil {
		logger.LogError(err)
		return infrastruct.ErrorInternalServerError
	}

	if err := s.p.DeleteSectionAndTeachersBDByTeacherID(idTeacher); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteSectionAndTeachersBD"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) GetStudentByID(idStudent int) (*types.User, error) {

	user, err := s.p.GetUserByID(idStudent)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(err)
			return nil, infrastruct.ErrorInternalServerError
		}
		return nil, infrastruct.ErrorNotFound
	}

	return user, nil
}

func (s *Service) GetUsers(role string) ([]types.UserStat, error) {

	moreUser := []types.UserStat{}

	var err error
	switch role {
	case "student":
		moreUser, err = s.p.GetAllStudentsForAdmin()
		if err != nil {
			logger.LogError(errors.Wrap(err, "Err with GetAllStudentsForAdmin"))
			return nil, infrastruct.ErrorInternalServerError
		}
	case "teacher":
		moreUser, err = s.p.GetAllTeachersForAdmin()
		if err != nil {
			logger.LogError(errors.Wrap(err, "Err with GetAllTeachersForAdmin"))
			return nil, infrastruct.ErrorInternalServerError
		}
	case "admin":
		moreUser, err = s.p.GetAllAdminsForAdmin()
		if err != nil {
			logger.LogError(errors.Wrap(err, "Err with GetAllAdminsForAdmin"))
			return nil, infrastruct.ErrorInternalServerError
		}
	case "":
		moreUser, err = s.p.GetAllUsersForAdmin()
		if err != nil {
			logger.LogError(errors.Wrap(err, "Err with GetAllUsersForAdmin"))
			return nil, infrastruct.ErrorInternalServerError
		}
	default:
		return nil, infrastruct.ErrorBadRequest
	}

	return moreUser, nil
}

func (s *Service) GetAllTeachersInfoForAdmin() ([]types.TeacherFullInfo, error) {

	teacherArr, err := s.p.GetAllTeachersInfoForAdmin()
	if err != nil {
		logger.LogError(errors.Wrap(err, "Err with GetAllTeachersInfoForAdmin"))
		return nil, infrastruct.ErrorInternalServerError
	}

	for i, _ := range teacherArr {
		teacherArr[i].Times = teacherArr[i].Times / 60 / 60
	}

	return teacherArr, nil
}

func (s *Service) GetAllCourse() ([]types.Course, error) {

	courseArr, err := s.p.GetAllCourse()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(errors.Wrap(err, "err with GetAllCourse"))
		return nil, infrastruct.ErrorInternalServerError
	}

	return courseArr, nil
}

func (s *Service) GetAllCoursesInfoForAdmin() ([]types.CourseInfoForAdmin, error) {

	courseArr, err := s.p.GetAllCoursesInfoForAdmin()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(errors.Wrap(err, "err with GetAllCoursesInfoForAdmin"))
		return nil, infrastruct.ErrorInternalServerError
	}

	return courseArr, nil
}

func (s *Service) GetAllSectionsInCourse(idCourse int) ([]types.Section, error) {

	//check idCourse in URL
	if err := s.p.CheckURLByC(idCourse); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	sectionArr, err := s.p.GetAllSectionInCourses(idCourse)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(errors.Wrap(err, "err with GetAllSectionInCourses"))
		return nil, infrastruct.ErrorInternalServerError
	}

	return sectionArr, nil
}

func (s *Service) GetAllLevelsInSection(idCourse, idSection int) ([]types.Level, error) {

	//check idCourse and idSection in URL
	if err := s.p.CheckURLByCS(idCourse, idSection); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	levelArr, err := s.p.GetAllLevelsInSection(idCourse, idSection)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(errors.Wrap(err, "err with GetAllLevelsInSection"))
		return nil, infrastruct.ErrorInternalServerError
	}

	return levelArr, nil
}

func (s *Service) GetAllLessonsInLevel(idCourse, idSection, idLevel int) ([]types.Lesson, error) {

	//check idCourse, idSection and idLevel in URL
	if err := s.p.CheckURLByCSL(idCourse, idSection, idLevel); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	lessonArr, err := s.p.GetAllLessonsInLevel(idCourse, idSection, idLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(errors.Wrap(err, "err with GetAllLessonsInLevel"))
		return nil, infrastruct.ErrorInternalServerError
	}

	return lessonArr, nil
}

func (s *Service) ChangePassword(ch *types.ChangePassword) error {
	ch.OldPassword = strings.TrimSpace(ch.OldPassword)
	ch.NewPassword = strings.TrimSpace(ch.NewPassword)
	ch.RepeatPassword = strings.TrimSpace(ch.RepeatPassword)

	if ch.NewPassword != ch.RepeatPassword {
		return infrastruct.ErrorPasswordsDoNotMatch
	}

	user, err := s.p.GetUserByID(ch.UserID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetUserByID"))
		return infrastruct.ErrorInternalServerError
	}

	if ch.OldPassword != user.Password {
		return infrastruct.ErrorPasswordIsIncorrect
	}

	if err := s.p.UpdatePassword(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdatePassword"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) RecoveryPassword(ch *types.RecoveryPasswordEmail) error {

	user, err := s.p.GetUserByEmail(ch.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetUserByID"))
			return infrastruct.ErrorInternalServerError
		}
		return infrastruct.ErrorEmailNotFind
	}

	//проверяем есть ли уже сгенерированный код - если есть, удаляем и делаем новый
	have, err := s.p.CheckDBRecoveryPassword(user.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with CheckDBRecoveryPassword"))
			return infrastruct.ErrorInternalServerError
		}
	}
	if have {
		if err := s.p.DeleteRecoveryPass(user.Email); err != nil {
			logger.LogError(errors.Wrap(err, "err with DeleteRecoveryPass"))
			return infrastruct.ErrorInternalServerError
		}
	}

	code, err := uuid.GenerateUUID()
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GenerateUUID"))
		return infrastruct.ErrorInternalServerError
	}
	code = code[:5]
	code = strings.ToUpper(code)

	if err := s.p.AddCodeForRecoveryPass(user.Email, code); err != nil {
		logger.LogError(errors.Wrap(err, "err with AddCodeForRecoveryPass"))
		return infrastruct.ErrorInternalServerError
	}

	r := mail.NewRequest([]string{user.Email}, s.email)
	if err = r.Send(s.htmlRecoveryPath, code); err != nil {
		logger.LogError(errors.Wrap(err, "err with send Email"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) CheckValidRecoveryPassword(ch *types.RecoveryPasswordEmailAndCode) (bool, error) {

	if err := s.p.CheckRecoveryPass(ch); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		logger.LogError(errors.Wrap(err, "err with CheckValidRecoveryPassword"))
		return false, infrastruct.ErrorInternalServerError
	}

	return true, nil
}

func (s *Service) NewRecoveryPassword(ch *types.RecoveryPasswordNewPass) error {
	ch.Code = strings.ToUpper(ch.Code)

	recovery := &types.RecoveryPasswordEmailAndCode{Email: ch.Email, Code: ch.Code}
	ok, err := s.CheckValidRecoveryPassword(recovery)
	if err != nil {
		return err
	}

	if !ok {
		//todo если пароль код и пасс не корректные - значит человек не прошел валидацию на предыдущем этапе, поэтому шлем внутреннюю ошибку
		return infrastruct.ErrorInternalServerError
	}

	if ch.NewPassword != ch.RepeatPassword {
		return infrastruct.ErrorPasswordsDoNotMatch
	}

	id, err := s.p.GetUserByEmail(ch.Email)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetUserByEmail"))
		return infrastruct.ErrorInternalServerError
	}

	newPass := &types.ChangePassword{UserID: id.ID, NewPassword: ch.NewPassword}
	if err := s.p.UpdatePassword(newPass); err != nil {
		logger.LogError(errors.Wrap(err, "err  with UpdatePass"))
		return infrastruct.ErrorInternalServerError
	}

	if err := s.p.DeleteRecoveryPass(recovery.Email); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteRecoveryPass"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) AddCourse(ch *types.Course) (*types.OnlyID, error) {

	ch.TotalPrice = ch.Cost
	id, err := s.p.CreateCourse(ch)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with CreateCourse"))
		return id, infrastruct.ErrorInternalServerError
	}

	return id, nil
}

func (s *Service) AddSection(ch *types.Section) (*types.OnlyID, error) {

	//check correct courseID in URL
	if err := s.p.CheckURLByC(ch.CourseID); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	id, err := s.p.CreateSection(ch)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with CreateSection"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		logger.LogError(err)
		return id, infrastruct.ErrorInternalServerError
	}

	return id, nil
}

func (s *Service) AddLevel(ch *types.Level) (*types.OnlyID, error) {

	//check correct courseID and sectionID in URL
	if err := s.p.CheckURLByCS(ch.CourseID, ch.SectionID); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	id, err := s.p.CreateLevel(ch)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with CreateLevel"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		return id, infrastruct.ErrorInternalServerError
	}

	return id, nil
}

func (s *Service) AddLesson(ch *types.Lesson) (*types.OnlyID, error) {

	//check correct courseID and sectionID in URL
	if err := s.p.CheckURLByCSL(ch.CourseID, ch.SectionID, ch.LevelID); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	id, err := s.p.CreateLesson(ch)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with CreateLesson"))
		return id, infrastruct.ErrorInternalServerError
	}

	carousel, err := s.p.GetLessonCarousel(ch.LevelID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetLessonCarousel"))
			return nil, infrastruct.ErrorInternalServerError
		}
		idLesson := []int{id.ID}
		if err := s.p.AddCarousel(ch.CourseID, ch.SectionID, ch.LevelID, idLesson); err != nil {
			logger.LogError(errors.Wrap(err, "err with AddCarousel"))
			return id, infrastruct.ErrorInternalServerError
		}
		return id, nil
	}

	carousel.LessonArray = append(carousel.LessonArray, int64(id.ID))
	if err = s.p.UpdateCarousel(carousel); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateCarousel"))
		return id, infrastruct.ErrorInternalServerError
	}

	return id, nil
}

func (s *Service) GetCourse(idCourse int) (*types.Course, error) {

	course, err := s.p.GetCourse(idCourse)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetCourse"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		return nil, infrastruct.ErrorInternalServerError
	}

	return course, nil
}

func (s *Service) GetSection(idCourse, idSection int) (*types.Section, error) {

	//check correct courseID in URL
	if err := s.p.CheckURLByCS(idCourse, idSection); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	section, err := s.p.GetSection(idSection)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetSection"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		return nil, infrastruct.ErrorInternalServerError
	}

	return section, nil
}

func (s *Service) GetLevel(idCourse, idSection, idLevel int) (*types.Level, error) {

	//check correct courseID and sectionID in URL
	if err := s.p.CheckURLByCSL(idCourse, idSection, idLevel); err != nil {
		return nil, infrastruct.ErrorNotFound
	}

	level, err := s.p.GetLevel(idLevel)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetLevel"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		return nil, infrastruct.ErrorInternalServerError
	}

	return level, nil
}

func (s *Service) GetLesson(idCourse, idSection, idLevel, idLesson int) (*types.Lesson, error) {

	//check correct courseID, sectionID and levelID in URL
	if err := s.p.CheckURLByCSLL(idCourse, idSection, idLevel, idLesson); err != nil {
		logger.LogError(errors.Wrap(err, "err with CheckURLByCSLL"))
		return nil, infrastruct.ErrorNotFound
	}

	lesson, err := s.p.GetLesson(idLesson)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetLesson"))
		if err == sql.ErrNoRows {
			return nil, infrastruct.ErrorNotFound
		}
		return nil, infrastruct.ErrorInternalServerError
	}

	//make video url
	arrLesson, err := s.p.GetLessonCarousel(idLevel)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetLessonCarousel"))
		return nil, infrastruct.ErrorInternalServerError
	}

	lesson.NextLessonID = 0
	if idLesson != int(arrLesson.LessonArray[len(arrLesson.LessonArray)-1]) {
		for i, v := range arrLesson.LessonArray {
			if int(v) == idLesson {
				lesson.NextLessonID = int(arrLesson.LessonArray[i+1])
				break
			}
		}
	}

	if lesson.NextLessonID != 0 {
		lesson.NextLessonURL = fmt.Sprintf("/courses/%d/sections/%d/levels/%d/courses/%d",
			lesson.CourseID, lesson.SectionID, lesson.LevelID, lesson.NextLessonID)
	}
	return lesson, nil
}

func (s *Service) UpdateCourse(ch *types.Course) error {

	//check correct courseID in URL
	if err := s.p.CheckURLByC(ch.ID); err != nil {
		return infrastruct.ErrorNotFound
	}

	//add sale
	if ch.Sale != 0 {
		if ch.Sale < 0 {
			ch.Sale = ch.Sale * (-1)
		}

		var sumSale float64

		sumSale = (float64(ch.Cost) * float64(ch.Sale)) / 100
		a := float64(ch.Cost) - sumSale
		ch.TotalPrice = int(a)
	} else {
		ch.TotalPrice = ch.Cost
	}

	if err := s.p.UpdateCourse(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateCourse"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) UpdateSection(ch *types.Section) error {

	//check correct courseID, sectionID in URL
	if err := s.p.CheckURLByCS(ch.CourseID, ch.ID); err != nil {
		return infrastruct.ErrorNotFound
	}

	if err := s.p.UpdateSection(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateSection"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) UpdateLevel(ch *types.Level) error {

	//check correct courseID and sectionID, levelID in URL
	if err := s.p.CheckURLByCSL(ch.CourseID, ch.SectionID, ch.ID); err != nil {
		return infrastruct.ErrorNotFound
	}

	if err := s.p.UpdateLevel(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateLevel"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) UpdateLesson(ch *types.Lesson) error {

	//check correct courseID, sectionID, levelID and LessonId in URL
	if err := s.p.CheckURLByCSLL(ch.CourseID, ch.SectionID, ch.LevelID, ch.ID); err != nil {
		return infrastruct.ErrorNotFound
	}

	if err := s.p.UpdateLesson(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateLesson"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) DeleteCourse(idCourse int) error {

	//check correct url
	if err := s.p.CheckURLByC(idCourse); err != nil {
		return infrastruct.ErrorNotFound
	}

	sections, err := s.GetAllSectionsInCourse(idCourse)
	if err != nil {
		return err
	}
	for index, _ := range sections {
		idSection := sections[index].ID
		if err := s.DeleteSection(idCourse, idSection); err != nil {
			if err != infrastruct.ErrorNotFound {
				return err
			}
		}
	}
	if err := s.p.DeleteCourse(idCourse); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteCourse"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) DeleteSection(idCourse, idSection int) error {

	//check correct url
	if err := s.p.CheckURLByCS(idCourse, idSection); err != nil {
		return infrastruct.ErrorNotFound
	}

	levels, err := s.GetAllLevelsInSection(idCourse, idSection)
	if err != nil {
		return err
	}

	for index, _ := range levels {
		idLevels := levels[index].ID
		if err := s.DeleteLevel(idCourse, idSection, idLevels); err != nil {
			if err != infrastruct.ErrorNotFound {
				return err
			}
		}
	}

	if err := s.p.DeleteSection(idCourse, idSection); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteSection"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	if err := s.p.DeleteSectionAndTeachersBDBySectionID(idSection); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteSectionAndTeachersBD"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) DeleteLevel(idCourse, idSection, idLevel int) error {

	//check correct url
	if err := s.p.CheckURLByCSL(idCourse, idSection, idLevel); err != nil {
		return infrastruct.ErrorNotFound
	}

	lessons, err := s.GetAllLessonsInLevel(idCourse, idSection, idLevel)
	if err != nil {
		return err
	}

	for index, _ := range lessons {
		idLesson := lessons[index].ID
		if err := s.DeleteLesson(idCourse, idSection, idLevel, idLesson); err != nil {
			if err != infrastruct.ErrorNotFound {
				return err
			}
		}
	}

	if err := s.p.DeleteLevel(idCourse, idSection, idLevel); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteLevel"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) DeleteLesson(idCourse, idSection, idLevel, idLesson int) error {

	//check correct url
	if err := s.p.CheckURLByCSLL(idCourse, idSection, idLevel, idLesson); err != nil {
		return infrastruct.ErrorNotFound
	}

	chatsIDArr, err := s.p.GetChatsIDByLessonID(idLesson)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetAllLevelsInSection"))
			return infrastruct.ErrorInternalServerError
		}
	}
	for index, _ := range chatsIDArr {
		if err := s.p.DeleteChat(chatsIDArr[index]); err != nil {
			if err != sql.ErrNoRows {
				logger.LogError(errors.Wrap(err, "err with DeleteChatByLessonID"))
				return infrastruct.ErrorInternalServerError
			}
		}
		if err := s.p.DeleteMessageByChatID(chatsIDArr[index]); err != nil {
			if err != sql.ErrNoRows {
				logger.LogError(errors.Wrap(err, "err with DeleteMessageByLessonID"))
				return infrastruct.ErrorInternalServerError
			}
		}
	}

	if err := s.p.DeleteLesson(idCourse, idSection, idLevel, idLesson); err != nil {
		logger.LogError(errors.Wrap(err, "err with DeleteLesson"))
		if err == sql.ErrNoRows {
			return infrastruct.ErrorNotFound
		}
		return infrastruct.ErrorInternalServerError
	}

	oldCarousel, err := s.p.GetLessonCarousel(idLevel)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetLessonCarousel"))
		return infrastruct.ErrorNotFound
	}

	newCarouselArray := make([]int64, 0)

	for _, v := range oldCarousel.LessonArray {
		if v == int64(idLesson) {
			continue
		}

		newCarouselArray = append(newCarouselArray, v)
	}

	newCarousel := &types.LessonCarousel{LessonArray: newCarouselArray, CourseID: idCourse,
		SectionID: idSection, LevelID: idLevel}
	if err = s.p.UpdateCarousel(newCarousel); err != nil {
		logger.LogError(errors.Wrap(err, "err with UpdateCarousel"))
		return infrastruct.ErrorInternalServerError
	}

	if len(newCarousel.LessonArray) == 0 {
		if err := s.p.DeleteCarousel(idLevel); err != nil {
			logger.LogError(errors.Wrap(err, "err with DeleteCarousel"))
			return infrastruct.ErrorInternalServerError
		}
	}

	return nil
}

func (s *Service) RecordTime(rec *types.RecordTime) error {

	lastSessionS, err := s.p.LastSession(rec.UserID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with s.p.LastSession"))
		}
	}

	lastSessionT, err := time.Parse(time.RFC3339, lastSessionS)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with time.Parse"))
	}

	then := time.Since(lastSessionT).Seconds()
	if int(then) < (30 * 60) {
		if err = s.p.AddTime(int(then), rec.UserID); err != nil {
			logger.LogError(errors.Wrap(err, "err with s.p.WriteTime()"))
		}
	}

	if err := s.p.RecordTime(rec); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetChatByLessonForStudent(chat *types.ChatData) (*types.ChatData, error) {

	var err error

	chat.ChatID, err = s.p.GetChatID(chat)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatID"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	if err = s.p.OffsetTeacherMessages(chat.ChatID); err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with offsetTeacherMessages"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chat.Messages, err = s.p.GetMessage(chat.ChatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetMessage"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	return chat, nil
}

func (s *Service) SendMessageToChatByLessonForStudent(chat *types.ChatData, mes *types.MessageBody) error {

	var err error

	mes.FirstName, err = s.p.GetUserNameByUserID(mes.UserID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetUserNameByUserID"))
		return infrastruct.ErrorInternalServerError
	}

	chat.ChatID, err = s.p.GetChatID(chat)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatID"))
			return infrastruct.ErrorInternalServerError
		}
	}

	if chat.ChatID == 0 {
		chat.ChatID, err = s.p.MakeChat(chat)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with MakeChat"))
			return infrastruct.ErrorInternalServerError
		}
	}

	if err = s.p.SendMessageChat(chat.ChatID, mes); err != nil {
		logger.LogError(errors.Wrap(err, "err with SendMessageChat"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) GetChatByProfileStudent(chatID int, claims *infrastruct.CustomClaims) (*types.ChatData, error) {

	var err error
	chat, err := s.p.GetChatDataByChatID(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		}
		return nil, infrastruct.ErrorNotFound
	}

	//проверка прав доступа к чату
	if claims.UserID != chat.StudentID {
		return nil, infrastruct.ErrorPermissionDenied
	}

	if err = s.p.OffsetTeacherMessages(chat.ChatID); err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with offsetTeacherMessages"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chat.Messages, err = s.p.GetMessage(chat.ChatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetMessage"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	return chat, nil
}

func (s *Service) SendMessageToChatByProfileStudent(chatID int, mes *types.MessageBody) error {

	var err error

	chat, err := s.p.GetChatDataByChatID(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		}
		return infrastruct.ErrorNotFound
	}

	//проверка прав доступа к чату
	if mes.UserID != chat.StudentID {
		return infrastruct.ErrorPermissionDenied
	}

	mes.FirstName, err = s.p.GetUserNameByUserID(mes.UserID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetUserNameByUserID"))
		return infrastruct.ErrorInternalServerError
	}

	if err = s.p.SendMessageChat(chat.ChatID, mes); err != nil {
		logger.LogError(errors.Wrap(err, "err with SendMessageChat"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) GetChatForTeacher(chatID int, claims *infrastruct.CustomClaims) (*types.ChatData, error) {

	var err error
	chat, err := s.p.GetChatDataByChatID(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		}
		return nil, infrastruct.ErrorNotFound
	}
	//todo нету проверки доступа к чату
	////проверка прав доступа к чату
	//if claims.UserID != chat.StudentID {
	//	return nil, infrastruct.ErrorPermissionDenied
	//}

	if err = s.p.OffsetStudentMessages(chat.ChatID); err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with offsetTeacherMessages"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chat.Messages, err = s.p.GetMessage(chat.ChatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetMessage"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	return chat, nil
}

func (s *Service) GetChatForAdmin(chatID int) (*types.ChatData, error) {

	var err error
	chat, err := s.p.GetChatDataByChatID(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		}
		return nil, infrastruct.ErrorNotFound
	}

	chat.Messages, err = s.p.GetMessage(chat.ChatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetMessage"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	return chat, nil
}

func (s *Service) SendMessageToChatForTeacher(chatID int, mes *types.MessageBody) error {

	var err error

	chat, err := s.p.GetChatDataByChatID(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		}
		return infrastruct.ErrorNotFound
	}
	//todo нету проверки доступа к чату
	////проверка прав доступа к чату
	//if mes.UserID != chat.StudentID {
	//	return infrastruct.ErrorPermissionDenied
	//}

	mes.FirstName, err = s.p.GetUserNameByUserID(mes.UserID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetUserNameByUserID"))
		return infrastruct.ErrorInternalServerError
	}

	timeLastStudentMesS, err := s.p.FindLastStudentMessageNotAnswer(chatID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with s.p.FindLastStudentMessageNotAnswer"))
		}
	}

	if timeLastStudentMesS != "" {
		timeLastStudentMesT, err := time.Parse(time.RFC3339, timeLastStudentMesS)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with time.Parse"))
		}

		then := time.Since(timeLastStudentMesT).Seconds()

		if err = s.p.AddTimeAnswer(mes.UserID, int(then)); err != nil {
			logger.LogError(errors.Wrap(err, "err with s.p.AddTimeAnswer"))
		}
	}

	if err = s.p.SendMessageChat(chat.ChatID, mes); err != nil {
		logger.LogError(errors.Wrap(err, "err with SendMessageChat"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) GetAllChatsForTeacher(teacherID int) ([]types.ChatsPreviewForTeacher, error) {

	//find id sections
	sectionsID, err := s.p.GetAllSectionsIDByTeacherID(teacherID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetAllSectionsIDByTeacherID"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chatsPreview := make([]types.ChatsPreviewForTeacher, 0)
	chatPreview := types.ChatsPreviewForTeacher{}

	for _, sectionID := range sectionsID {
		chatsID, err := s.p.GetChatsIDBySectionID(sectionID)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with GetChatsIDBySectionID"))
			return nil, infrastruct.ErrorInternalServerError
		}
		for _, chatID := range chatsID {
			chatPreview.ChatID = chatID
			chatPreview.SectionsID = sectionID
			chatsPreview = append(chatsPreview, chatPreview)
		}
	}

	wg := sync.WaitGroup{}
	errorChan := make(chan struct{}, len(chatsPreview))
	defer close(errorChan)

	for index, _ := range chatsPreview {
		wg.Add(1)
		go func(index int) {
			wg.Done()
			chatsPreview[index].StudentID, err = s.p.GetStudentIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetStudentIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].StudentFirstName, err = s.p.GetUserNameByUserID(chatsPreview[index].StudentID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetUserNameByUserID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonID, err = s.p.GetLessonIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonName, err = s.p.GetLessonNameByLessonID(chatsPreview[index].LessonID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonNameByLessonID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].SectionsName, err = s.p.GetSectionNameBySectionsID(chatsPreview[index].SectionsID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetSectionNameBySectionsID"))
				errorChan <- struct{}{}
				return
			}
			timeS, err := s.p.GetTimeByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetTimeByChatID"))
				errorChan <- struct{}{}
				return
			}
			timeT, err := time.Parse(time.RFC3339, timeS)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with time.Parse"))
				errorChan <- struct{}{}
				return
			}
			timeT = timeT.Add(time.Hour * 24)
			chatsPreview[index].Time = timeT.Format(time.RFC3339)

			chatsPreview[index].Ahtung, err = s.p.GetAhtungByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetAhtungByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].NotViewMessage, err = s.p.GetNotViewMessageForTeacherByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetAhtungByChatID"))
				errorChan <- struct{}{}
				return
			}
		}(index)
	}
	wg.Wait()
	select {
	case <-errorChan:
		return nil, infrastruct.ErrorInternalServerError
	default:
		return chatsPreview, nil
	}
}

func (s *Service) GetAllChatsForAdmin(teacherID int) ([]types.ChatsPreviewForAdmin, error) {
	//find id sections
	sectionsID, err := s.p.GetAllSectionsIDByTeacherID(teacherID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetAllSectionsIDByTeacherID"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chatsPreview := make([]types.ChatsPreviewForAdmin, 0)
	chatPreview := types.ChatsPreviewForAdmin{}

	for _, sectionID := range sectionsID {
		chatsID, err := s.p.GetChatsIDBySectionID(sectionID)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with GetChatsIDBySectionID"))
			return nil, infrastruct.ErrorInternalServerError
		}
		for _, chatID := range chatsID {
			chatPreview.ChatID = chatID
			chatPreview.SectionsID = sectionID
			chatsPreview = append(chatsPreview, chatPreview)
		}
	}

	wg := &sync.WaitGroup{}
	errorChan := make(chan struct{}, len(chatsPreview))
	defer close(errorChan)

	for index, _ := range chatsPreview {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			chatsPreview[index].StudentID, err = s.p.GetStudentIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetStudentIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].StudentFirstName, err = s.p.GetUserNameByUserID(chatsPreview[index].StudentID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetUserNameByUserID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonID, err = s.p.GetLessonIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonName, err = s.p.GetLessonNameByLessonID(chatsPreview[index].LessonID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonNameByLessonID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].SectionsName, err = s.p.GetSectionNameBySectionsID(chatsPreview[index].SectionsID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetSectionNameBySectionsID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].Time, err = s.p.GetTimeByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetTimeByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].Rating, err = s.p.GetRatingByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetRatingByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LastMessage, err = s.p.GetLastMessageByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLastMessageByChatID"))
				errorChan <- struct{}{}
				return
			}
		}(index)
	}

	wg.Wait()
	select {
	case <-errorChan:
		return nil, infrastruct.ErrorInternalServerError
	default:
		return chatsPreview, nil
	}
}

func (s *Service) GetAllChatsForStudent(studentID int) ([]types.ChatsPreviewForStudent, error) {
	//find id sections
	chatsID, err := s.p.GetAllChatsIDByStudentID(studentID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with GetAllSectionsIDByTeacherID"))
			return nil, infrastruct.ErrorInternalServerError
		}
	}

	chatsPreview := make([]types.ChatsPreviewForStudent, 0)
	chatPreview := types.ChatsPreviewForStudent{}

	for _, chatID := range chatsID {
		chatPreview.ChatID = chatID
		chatsPreview = append(chatsPreview, chatPreview)
	}

	wg := &sync.WaitGroup{}
	errorChan := make(chan struct{}, len(chatsPreview))
	defer close(errorChan)

	for index, _ := range chatsPreview {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			chatsPreview[index].SectionsID, err = s.p.GetSectionIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetSectionIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].SectionsName, err = s.p.GetSectionNameBySectionsID(chatsPreview[index].SectionsID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetSectionNameBySectionsID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonID, err = s.p.GetLessonIDByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonIDByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LessonName, err = s.p.GetLessonNameByLessonID(chatsPreview[index].LessonID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLessonNameByLessonID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].LastMessage, err = s.p.GetLastTeacherMessageByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetLastTeacherMessageByChatID"))
				errorChan <- struct{}{}
				return
			}
			chatsPreview[index].NotViewMessage, err = s.p.GetNotViewMessageForStudentByChatID(chatsPreview[index].ChatID)
			if err != nil {
				logger.LogError(errors.Wrap(err, "err with GetNotViewMessageForStudentByChatID"))
				errorChan <- struct{}{}
				return
			}
		}(index)
	}
	wg.Wait()
	select {
	case <-errorChan:
		return nil, infrastruct.ErrorInternalServerError
	default:
		return chatsPreview, nil
	}
}

func (s *Service) Ahtung(ch *types.Ahtung) error {
	//todo нету проверки доступа к чату
	ahtung, err := s.p.GetAhtungByChatID(ch.ChatID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetAhtungByChatID"))
		return infrastruct.ErrorInternalServerError
	}
	if ahtung {
		ch.Ahtung = false
		if err := s.p.ChangeAhtung(ch); err != nil {
			logger.LogError(errors.Wrap(err, "err with ChangeAhtung"))
			return infrastruct.ErrorInternalServerError
		}
	} else {
		ch.Ahtung = true
		if err := s.p.ChangeAhtung(ch); err != nil {
			logger.LogError(errors.Wrap(err, "err with ChangeAhtung"))
			return infrastruct.ErrorInternalServerError
		}

		if err = s.p.IncrementAhtung(ch.TeacherID); err != nil {
			logger.LogError(errors.Wrap(err, "err with IncrementAhtung"))
		}
	}

	return nil
}

func (s *Service) Rating(ch *types.Rating) error {
	//todo нету проверки доступа к чату
	if ch.Rating != "good" && ch.Rating != "improve" {
		return infrastruct.ErrorBadRequest
	}

	if ch.Rating == "good" {
		if err := s.p.IncrementGood(ch.TeacherID); err != nil {
			logger.LogError(errors.Wrap(err, "err with IncrementGood"))
			return infrastruct.ErrorInternalServerError
		}
	} else {
		if err := s.p.IncrementImprove(ch.TeacherID); err != nil {
			logger.LogError(errors.Wrap(err, "err with IncrementImprove"))
			return infrastruct.ErrorInternalServerError
		}
	}

	oldRating, err := s.p.GetRatingByChatID(ch.ChatID)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with GetRatingByChatID"))
		return infrastruct.ErrorInternalServerError
	}

	if oldRating == "" {
		courseID, err := s.p.GetCourseIDByChatID(ch.ChatID)
		if err != nil {
			logger.LogError(errors.Wrap(err, "err with GetCourseIDByChatID"))
			return infrastruct.ErrorInternalServerError
		}
		if err = s.p.IncrementHomeWork(courseID); err != nil {
			logger.LogError(errors.Wrap(err, "err with IncrementHomeWork"))
			return infrastruct.ErrorInternalServerError
		}
	}

	if err := s.p.ChangeRating(ch); err != nil {
		logger.LogError(errors.Wrap(err, "err with GetChatDataByChatID"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) UploadVideo(video *types.UploadVideo) error {

	//check url
	if err := s.p.CheckURLByCSLL(video.CourseID, video.SectionID, video.LevelID, video.LessonID); err != nil {
		return infrastruct.ErrorNotFound
	}

	//Проверить наличие видео
	existenceVideoID, err := s.checkExistenceVideo(video.LessonID)
	if err != nil {
		return infrastruct.ErrorInternalServerError
	}
	if existenceVideoID > 0 {
		if err := os.Remove(fmt.Sprintf("%s/%d", s.videoDir, existenceVideoID)); err != nil {
			logger.LogError(errors.Wrap(err, "err with os.Remove"))
			return infrastruct.ErrorInternalServerError
		}
	}

	dst, err := os.Create(filepath.Join(s.videoDir, fmt.Sprintf("%d", video.LessonID)))
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with Create in UploadVideo"))
		return infrastruct.ErrorInternalServerError
	}

	_, err = io.Copy(dst, video.Body)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with Copy in UploadVideo"))
		return infrastruct.ErrorInternalServerError
	}

	return nil
}

func (s *Service) CheckUserInDBUsers(userID int) error {

	if err := s.p.CheckUserInDBUsers(userID); err != nil {
		if err != sql.ErrNoRows {
			logger.LogError(errors.Wrap(err, "err with CheckUserInDBUsers"))
			return infrastruct.ErrorInternalServerError
		}
		return infrastruct.ErrorPermissionDenied
	}

	return nil
}

func trimSpaceUser(user *types.User) {
	user.Password = strings.TrimSpace(user.Password)
	user.FirstName = strings.TrimSpace(user.FirstName)
	user.Email = strings.TrimSpace(user.Email)
}

func trimSpaceTeacher(teacher *types.Teacher) {
	teacher.Password = strings.TrimSpace(teacher.Password)
	teacher.FirstName = strings.TrimSpace(teacher.FirstName)
	teacher.Email = strings.TrimSpace(teacher.Email)
}

func (s *Service) checkExistenceVideo(lessonID int) (int, error) {

	fileInfo, err := os.Stat(fmt.Sprintf("%s/%d", s.videoDir, lessonID))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		logger.LogError(errors.Wrap(err, "err with os.IsNotExist"))
		return 0, err
	}
	name, err := strconv.Atoi(fileInfo.Name())
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with strconv.Atoi"))
		return 0, err
	}

	return name, nil
}
