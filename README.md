# hcr

使用 goharbor-client 库 发送habor restful api 


https://unioslo.github.io/harborapi/reference/client/#harborapi.client.HarborAsyncClient.delete_registry


https://v3-1.docs.kubesphere.io/zh/docs/devops-user-guide/how-to-integrate/harbor/#%E8%8E%B7%E5%8F%96-harbor-%E5%87%AD%E8%AF%81


echo https://github.com/labstack/echo/discussions/2158

1. 凭证： https://v3-1.docs.kubesphere.io/zh/docs/devops-user-guide/how-to-integrate/harbor/#%E8%8E%B7%E5%8F%96-harbor-%E5%87%AD%E8%AF%81

2. NewRegistry 是干啥用的？ RegistryCredential 是凭证，是用来配置replication的。 

3. 凭证，就是 AddProjectRobotV1

4. Artifacts 参见： https://www.sohu.com/a/433563325_609552



## to study 

github.com/go-openapi


[put和PATCH](https://segmentfault.com/q/1010000005685904)

```go
type CreateTagParams struct {

	/*XRequestID
	  An unique ID for the request

	*/
	XRequestID *string
	/*ProjectName
	  The name of the project

	*/
	ProjectName string
	/*Reference
	  The reference of the artifact, can be digest or tag

	*/
	Reference string
	/*RepositoryName
	  The name of the repository. If it contains slash, encode it with URL encoding. e.g. a/b -> a%252Fb

	*/
	RepositoryName string
	/*Tag
	  The JSON object of tag.

	*/
	Tag *model.Tag

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

*/