package main

import (
	"context"

	"net/http"
	"os"

	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	harbor "github.com/mittwald/goharbor-client/v5/apiv2"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	HarborUrl string
	Username  string
	Password  string
}

func NewProject(c echo.Context) error {

	ac := c.(*AppContext)

	type Request struct {
		ProjectName string `json:"project_name" validate:"required"`
		UserName    string `json:"user_name"   validate:"required"`
	}

	r := new(Request)
	err := ac.validateAndBindRequest(r)
	if err != nil {
		return err
	}

	projectReq := new(modelv2.ProjectReq)
	projectReq.ProjectName = r.ProjectName
	projectReq.Metadata.Public = "false" //默认为私有仓库

	//request.Validate() 如何使用

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Fatalf("NewRESTClientForHost")
	}

	err = client.NewProject(context.Background(), projectReq)
	if err != nil {
		log.Errorf("failed to NewProject: " + err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorRes{Code: "901", Detail: err.Error()})
	}

	m := new(modelv2.ProjectMember)
	m.RoleID = 1 //Admin
	m.MemberUser.Username = r.UserName

	err = client.AddProjectMember(context.Background(), r.ProjectName, m)
	if err != nil {
		log.Errorf("failed to AddProjectMember: " + err.Error())
		return c.JSON(http.StatusInternalServerError, ErrorRes{Code: "902", Detail: err.Error()})
	}

	return nil

}

func NewProjectCredential(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	projectName := c.Param("project")

	type ProjectCredential struct {
		Name        string `json:"name"`
		Password    string `json:"password"`
		ProjectName string `json:"project_name"`
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("NewRESTClientForHost")
		return err
	}

	rc := new(modelv2.RobotCreate)
	rc.Name = "robot1"

	rs, err := client.NewRobotAccount(ctx, rc)
	if err != nil {
		log.Errorf("NewRESTClientForHost")
		return err
	}

	pc := new(ProjectCredential)
	pc.Name = rs.Name
	pc.Password = rs.Secret
	pc.ProjectName = projectName

	m := new(modelv2.ProjectMember)
	m.RoleID = 2
	m.MemberUser.Username = rs.Name

	err = client.AddProjectMember(ctx, projectName, m)
	if err != nil {
		log.Errorf("AddProjectMember")
		client.DeleteRobotAccountByName(ctx, rc.Name)
		return err
	}

	return ac.okResponseWithData(pc)

}

func DelProject(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	name := c.Param("project")

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, nil)
	if err != nil {
		log.Fatalf("NewRESTClientForHost")
	}

	client.DeleteProject(ctx, name)

	return nil
}

func QueryUserIsProjectMember(client *harbor.RESTClient, username string, project *modelv2.Project) (uint8, error) {
	ctx := context.Background()

	members, err := client.ListProjectMembers(ctx, project.Name, "")
	if err != nil {
		return 0, err
	}

	for _, v := range members {
		if v.EntityName == username {
			return 1, err
		}
	}

	return 0, err
}

func ListProjects(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	type Response struct {
		Projects []string `json:"projects"`
	}

	username := c.Param("user")
	result_projects := make([]string, 5)

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("NewRESTClientForHost")
		return err
	}

	projects, err := client.ListProjects(ctx, "")

	for _, project := range projects {
		res, err := QueryUserIsProjectMember(client, username, project)
		if err != nil {
			log.Errorf("failed to QueryUserIsProjectMember")
			return err
		}

		if res > 0 {
			result_projects = append(result_projects, project.Name)
		}
	}

	return ac.okResponseWithData(Response{Projects: result_projects})

}

func NewUser(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	type NewUserRequest struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required"`
		Realname string `json:"readname" validate:"required"`
		Password string `json:"password" validate:"required"`
		Comments string `json:"comments" validate:"required"`
	}

	r := new(NewUserRequest)
	err := ac.validateAndBindRequest(r)
	if err != nil {
		log.Errorf("failed to validateAndBindRequest")
		return err
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	err = client.NewUser(ctx, r.Username, r.Email, r.Realname, r.Password, r.Comments)
	if err != nil {
		log.Errorf("failed to NewUser")
		return err
	}

	return nil

}

func DelUser(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	username := c.Param("user")

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	userResp, err := client.GetUserByName(ctx, username)
	if err != nil {
		log.Errorf("failed to GetUserByName")
		return err
	}

	err = client.DeleteUser(ctx, userResp.UserID)
	if err != nil {
		log.Errorf("failed to DeleteUser")
		return err
	}

	return nil

}

func UpdateUserProfile(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	type UserProfileReq struct {
		Username string `json:"username"`
		Comment  string `json:"comment,omitempty"`
		Email    string `json:"email,omitempty"`
		Realname string `json:"realname,omitempty"`
	}

	r := new(UserProfileReq)

	err := ac.validateAndBindRequest(r)
	if err != nil {
		log.Errorf("failed to validateAndBindRequest")
		return err
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	userResp, err := client.GetUserByName(ctx, r.Username)
	if err != nil {
		log.Errorf("failed to GetUserByName")
		return err
	}

	profile := new(modelv2.UserProfile)
	profile.Comment = r.Comment
	profile.Email = r.Email
	profile.Realname = r.Realname

	err = client.UpdateUserProfile(ctx, userResp.UserID, profile)
	if err != nil {
		log.Errorf("failed to UpdateUserProfile")
		return err
	}

	return nil

}

func UpdateUserPassword(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	type UserPasswordRequest struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	r := new(UserPasswordRequest)

	err := ac.validateAndBindRequest(r)
	if err != nil {
		log.Errorf("failed to validateAndBindRequest")
		return err
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, config.Defaults())
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	userResp, err := client.GetUserByName(ctx, r.Username)
	if err != nil {
		log.Errorf("failed to GetUserByName")
		return err
	}

	passReq := new(modelv2.PasswordReq)
	passReq.NewPassword = r.NewPassword
	passReq.OldPassword = r.OldPassword

	err = client.UpdateUserPassword(ctx, userResp.UserID, passReq)
	if err != nil {
		log.Errorf("failed to UpdateUserPassword")
		return err
	}

	return ac.okResponse()

}

func ListRepositories(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	projectName := c.Param("project")

	result_repositories := make([]string, 1)
	type Response struct {
		Images []string `json:"images"`
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, nil)
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	reps, err := client.ListRepositories(ctx, projectName)
	if err != nil {
		log.Errorf("failed to ListRepositories")
		return err
	}

	for _, rep := range reps {
		result_repositories = append(result_repositories, rep.Name)
	}

	return ac.okResponseWithData(Response{Images: result_repositories})
}

func ListImageTags(c echo.Context) error {
	ac := c.(*AppContext)
	ctx := context.Background()

	projectName := c.Param("project")
	repo := c.Param("repository")

	tag_list := make([]string, 1)

	type Response struct {
		Project     string   `json:"project"`
		Respository string   `json:"repository"`
		Tags        []string `json:"tags"`
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, nil)
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	tagList, err := client.ListTags(ctx, projectName, repo, "")
	for _, tag := range tagList {
		tag_list = append(tag_list, tag.Name)
	}

	return ac.okResponseWithData(Response{Project: projectName, Respository: repo, Tags: tag_list})
}

func DeleteRepository(c echo.Context) error {

	ac := c.(*AppContext)
	ctx := context.Background()

	projectName := c.Param("project")
	imageName := c.Param("repository")

	type Response struct {
		Images []string `json:"images"`
	}

	client, err := harbor.NewRESTClientForHost(ac.Config.HarborUrl, ac.Config.Username, ac.Config.Password, nil)
	if err != nil {
		log.Errorf("failed to NewRESTClientForHost")
		return err
	}

	err = client.DeleteRepository(ctx, projectName, imageName)
	if err != nil {
		log.Errorf("failed to DeleteRepository")
		return err
	}

	return ac.okResponse()
}

func main() {

	config := new(Config)

	config.HarborUrl = os.Getenv("HARBORURL")
	if config.HarborUrl != "" {
		log.Fatalf("env HARBORURL is null")
	}

	config.Username = os.Getenv("HARBOR_ADMIN")
	if config.HarborUrl != "" {
		log.Fatalf("env HARBOR_ADMIN is null")
	}

	config.Password = os.Getenv("HARBOR_PASSWORD")
	if config.HarborUrl != "" {
		log.Fatalf("env HARBOR_PASSWORD is null")
	}

	e := echo.New()

	e.Logger.SetLevel(log.DEBUG)

	e.Use(echoPrometheus.MetricsMiddleware())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			customContext := &AppContext{
				Context: c,
				Config:  *config,
			}
			return next(customContext)
		}
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	api := e.Group("/api/v1")

	api.POST("/project/", NewProject)
	api.DELETE("/project/:project", DelProject)

	//获取某个用户的Project列表
	api.GET("/project/:user", ListProjects)

	//获取某个project下的镜像列表
	api.GET("/project/:project/repositories", ListRepositories)
	api.DELETE("/project/:project/repository/:repository", DeleteRepository)

	//获取某个repository的tag列表
	api.GET("/project/:project/repository/:repository/tags", ListImageTags) //ListArtifacts,ListTags
	//ListTags

	//用户
	api.POST("/user", NewUser)
	api.DELETE("/user/:user", DelUser)

	//创建一个project的凭证
	api.POST("/project/:project/credential", NewProjectCredential) //AddProjectRobotV1， NewRobotAccount-》AddProjectMember

	api.PATCH("/userpassword", UpdateUserPassword) //修改用户密码
	api.PATCH("/userprofile", UpdateUserProfile)   //修改用户profile

	e.Logger.Fatal(e.Start(":4001"))
}
