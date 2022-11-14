package generate

import (
	"errors"
	"fmt"

	"github.com/zhufuyi/sponge/pkg/replacer"

	"github.com/huandu/xstrings"
	"github.com/spf13/cobra"
)

// RPCGwPbCommand generate rpc gateway server codes base on protobuf file
func RPCGwPbCommand() *cobra.Command {
	var (
		moduleName   string // go.mod文件的module名称
		serverName   string // 服务名称
		projectName  string // 项目名称
		repoAddr     string // 镜像仓库地址
		outPath      string // 输出目录
		protobufFile string // proto file文件
	)

	//nolint
	cmd := &cobra.Command{
		Use:   "rpc-gw-pb",
		Short: "Generate rpc gateway server codes based on protobuf file",
		Long: `generate rpc gateway server codes based on protobuf file.

Examples:
  # generate rpc gateway server codes.
  sponge micro rpc-gw-pb --module-name=yourModuleName --server-name=yourServerName --project-name=yourProjectName --protobuf-file=./demo.proto

  # generate rpc gateway server codes and specify the output directory, Note: if the file already exists, code generation will be canceled.
  sponge micro rpc-gw-pb --module-name=yourModuleName --server-name=yourServerName --project-name=yourProjectName --protobuf-file=./demo.proto --out=./yourServerDir

  # generate rpc gateway server codes and specify the docker image repository address.
  sponge micro rpc-gw-pb --module-name=yourModuleName --server-name=yourServerName --project-name=yourProjectName --repo-addr=192.168.3.37:9443/user-name --protobuf-file=./demo.proto
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenRPCGwCommand(moduleName, serverName, projectName, protobufFile, repoAddr, outPath)
		},
	}

	cmd.Flags().StringVarP(&moduleName, "module-name", "m", "", "module-name is the name of the module in the 'go.mod' file")
	_ = cmd.MarkFlagRequired("module-name")
	cmd.Flags().StringVarP(&serverName, "server-name", "s", "", "server name")
	_ = cmd.MarkFlagRequired("server-name")
	cmd.Flags().StringVarP(&projectName, "project-name", "p", "", "project name")
	_ = cmd.MarkFlagRequired("project-name")
	cmd.Flags().StringVarP(&protobufFile, "protobuf-file", "f", "", "proto file")
	_ = cmd.MarkFlagRequired("protobuf-file")

	cmd.Flags().StringVarP(&repoAddr, "repo-addr", "r", "", "docker image repository address, excluding http and repository names")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "output directory, default is ./serverName_rpc-gw-pb_<time>")

	return cmd
}

func runGenRPCGwCommand(moduleName string, serverName string, projectName string, protobufFile string, repoAddr string, outPath string) error {
	protobufFiles, isImportTypes, err := parseProtobufFiles(protobufFile)
	if err != nil {
		return err
	}

	subTplName := "rpc-gw-pb"
	r := Replacers[TplNameSponge]
	if r == nil {
		return errors.New("replacer is nil")
	}

	// 设置模板信息
	subDirs := []string{ // 只处理的子目录
		"api/types", "cmd/serverNameExample_grpcGwPbExample",
		"sponge/build", "sponge/configs", "sponge/deployments", "sponge/docs", "sponge/scripts", "sponge/third_party",
		"internal/config", "internal/ecode", "internal/routers", "internal/rpcclient", "internal/server",
	}
	subFiles := []string{ // 只处理子文件
		"sponge/.gitignore", "sponge/.golangci.yml", "sponge/go.mod", "sponge/go.sum",
		"sponge/Jenkinsfile", "sponge/Makefile", "sponge/README.md",
	}
	ignoreDirs := []string{} // 指定子目录下忽略处理的目录
	ignoreFiles := []string{ // 指定子目录下忽略处理的文件
		"types.pb.validate.go", "types.pb.go", // api/types
		"swagger.json", "swagger.yaml", "apis.swagger.json", "apis.html", "docs.go", // sponge/docs
		"userExample_rpc.go", "systemCode_http.go", "userExample_http.go", // internal/ecode
		"routers.go", "routers_test.go", "userExample.go", "userExample_service.pb.go", // internal/routers
		"grpc.go", "grpc_option.go", "grpc_test.go", // internal/server
	}

	if !isImportTypes {
		ignoreFiles = append(ignoreFiles, "types.proto")
	}

	r.SetSubDirsAndFiles(subDirs, subFiles...)
	r.SetIgnoreSubDirs(ignoreDirs...)
	r.SetIgnoreSubFiles(ignoreFiles...)
	fields := addRPCGwFields(moduleName, serverName, projectName, repoAddr, r)
	r.SetReplacementFields(fields)
	_ = r.SetOutputDir(outPath, serverName+"_"+subTplName)
	if err = r.SaveFiles(); err != nil {
		return err
	}

	_ = saveProtobufFiles(moduleName, serverName, r.GetOutputDir(), protobufFiles)
	_ = saveGenInfo(moduleName, serverName, r.GetOutputDir())

	fmt.Printf("generate %s's rpc gateway server codes successfully, out = %s\n\n", serverName, r.GetOutputDir())

	return nil
}

func addRPCGwFields(moduleName string, serverName string, projectName string, repoAddr string,
	r replacer.Replacer) []replacer.Field {
	var fields []replacer.Field

	repoHost, _ := parseImageRepoAddr(repoAddr)

	fields = append(fields, deleteFieldsMark(r, httpFile, startMark, endMark)...)
	fields = append(fields, deleteFieldsMark(r, dockerFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, dockerFileBuild, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, dockerComposeFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, k8sDeploymentFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, k8sServiceFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, imageBuildFile, wellOnlyGrpcStartMark, wellOnlyGrpcEndMark)...)
	fields = append(fields, deleteFieldsMark(r, makeFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, gitIgnoreFile, wellStartMark, wellEndMark)...)
	fields = append(fields, deleteFieldsMark(r, protoShellFile, wellStartMark, wellEndMark)...)
	fields = append(fields, replaceFileContentMark(r, readmeFile, "## "+serverName)...)
	fields = append(fields, []replacer.Field{
		{ // 替换Dockerfile文件内容
			Old: dockerFileMark,
			New: dockerFileHTTPCode,
		},
		{ // 替换Dockerfile_build文件内容
			Old: dockerFileBuildMark,
			New: dockerFileBuildHTTPCode,
		},
		{ // 替换docker-compose.yml文件内容
			Old: dockerComposeFileMark,
			New: dockerComposeFileHTTPCode,
		},
		{ // 替换*-deployment.yml文件内容
			Old: k8sDeploymentFileMark,
			New: k8sDeploymentFileHTTPCode,
		},
		{ // 替换*-svc.yml文件内容
			Old: k8sServiceFileMark,
			New: k8sServiceFileHTTPCode,
		},
		{ // 替换proto.sh文件内容
			Old: protoShellFileMark,
			New: protoShellServiceCode,
		},
		{
			Old: "github.com/zhufuyi/sponge",
			New: moduleName,
		},
		{
			Old: moduleName + "/pkg",
			New: "github.com/zhufuyi/sponge/pkg",
		},
		{
			Old: "sponge api docs",
			New: serverName + " api docs",
		},
		{
			Old: "serverNameExample",
			New: serverName,
		},
		// docker镜像和k8s部署脚本替换
		{
			Old: "server-name-example",
			New: xstrings.ToKebabCase(serverName),
		},
		{
			Old: "projectNameExample",
			New: projectName,
		},
		// docker镜像和k8s部署脚本替换
		{
			Old: "project-name-example",
			New: xstrings.ToKebabCase(projectName),
		},
		{
			Old: "repo-addr-example",
			New: repoAddr,
		},
		{
			Old: "image-repo-host",
			New: repoHost,
		},
		{
			Old: "_grpcGwPbExample",
			New: "",
		},
		{
			Old: "_mixExample",
			New: "",
		},
		{
			Old: "_pbExample",
			New: "",
		},
	}...)

	return fields
}