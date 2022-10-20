package middleware

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const codeTpl = `package ${package}

import (
	"encoding/base64"
	"github.com/wenlaizhou/middleware"
)

func init() {
	${code}
}
`

const varTpl = "var ${name}, _ = base64.StdEncoding.DecodeString(`${value}`)\n" + `
	middleware.RegisterHandler("${path}", func(context middleware.Context) {
		context.OK("${contentType}", ${name})
	})`

type distVar struct {
	Name        string
	Value       string
	Path        string
	ContentType string
}

type pageCode struct {
	Package string
	Code    string
}

// DistFrontend2Code 将前端编译代码即静态资源直接编译为go代码
//
//
// pkg: 包名 package ${pkg}
//
// distPath: 编译代码所在路径
//
// urlPrefix: 前端请求路径前缀, 默认为 /
func DistFrontend2Code(pkg string, distPath string, urlPrefix string) (string, error) {

	urlPrefix = strings.TrimSpace(urlPrefix)
	if len(urlPrefix) <= 0 {
		urlPrefix = "/"
	} else {
		if !strings.HasPrefix(urlPrefix, "/") {
			urlPrefix = fmt.Sprintf("/%v", urlPrefix)
		}
		if !strings.HasSuffix(urlPrefix, "/") {
			urlPrefix = fmt.Sprintf("%v/", urlPrefix)
		}
	}

	pkg = strings.TrimSpace(pkg)

	distPath = strings.TrimSpace(distPath)
	if HasEmptyString(pkg, distPath) {
		return "", errors.New("参数不全")
	}

	page := pageCode{
		Package: pkg,
	}
	codeBuilder := strings.Builder{}

	if err := filepath.Walk(distPath, func(path string, info fs.FileInfo, err error) error {
		fmt.Printf("开始编译: %v\n", path)
		if info == nil {
			return nil
		}
		path = strings.TrimSpace(path)
		newPath := strings.ReplaceAll(path, string(os.PathSeparator), "/")
		subPath := strings.Replace(newPath, fmt.Sprintf("%v/", distPath), "", 1)
		if info.IsDir() {
			if indexContent, err := os.ReadFile(fmt.Sprintf("%v%vindex.html", path, string(os.PathSeparator))); err == nil {
				name := fmt.Sprintf("var_rand_%v", rand.Int()) + strings.ReplaceAll(strings.ReplaceAll(info.Name(), ".", "_"), "-", "_")
				codeBuilder.WriteString(StringFormatStructs(varTpl, distVar{
					Name:        name,
					Value:       base64.StdEncoding.EncodeToString(indexContent),
					Path:        fmt.Sprintf("%v%v", urlPrefix, subPath),
					ContentType: Html,
				}))
				indexCode := fmt.Sprintf(`
	middleware.RegisterHandler("%v%v" ,func(context middleware.Context) {
		context.AddCacheHeader(3600 * 24 * 30)
		context.OK(middleware.Html, %v)
	})
`, urlPrefix, subPath, name)
				indexCode2 := fmt.Sprintf(`
	middleware.RegisterHandler("%v%v/index.html" ,func(context middleware.Context) {
		context.AddCacheHeader(3600 * 24 * 30)
		context.OK(middleware.Html, %v)
	})
`, urlPrefix, subPath, name)
				codeBuilder.WriteString(indexCode)
				codeBuilder.WriteString(indexCode2)
			} else {
				println(path)
				println("没有index.html")
				println(err.Error())
			}

			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			println(err.Error())
			return nil
		}

		name := fmt.Sprintf("var_rand_%v", rand.Int()) +
			strings.ReplaceAll(strings.ReplaceAll(info.Name(), ".", "_"), "-", "_")
		if subPath == "index.html" {
			name = "index_html"
		}
		url := fmt.Sprintf("%v%v", urlPrefix, subPath)
		contentType := http.DetectContentType(content)
		switch Ext(path) {
		case ".js":
		case ".jsx":
			contentType = Js
			break
		case ".css":
			contentType = Css
			break
		case ".json":
			contentType = Json
			break
		case ".html":
			contentType = Html
			break
		}
		codeBuilder.WriteString(StringFormatStructs(varTpl, distVar{
			Name:        name,
			Value:       base64.StdEncoding.EncodeToString(content),
			Path:        url,
			ContentType: contentType,
		}))

		codeBuilder.WriteString("\n")
		return nil
	}); err != nil {
		println(err.Error())
	}
	if urlPrefix == "/" {
		codeBuilder.WriteString(`
	middleware.RegisterIndex(func(context middleware.Context) {
		context.AddCacheHeader(3600 * 24 * 30)
		context.OK(middleware.Html, index_html)
	})
`)
	} else {
		codeBuilder.WriteString(fmt.Sprintf(`
	middleware.RegisterHandler("%v", func(context middleware.Context) {
		context.AddCacheHeader(3600 * 24 * 30)
		context.OK(middleware.Html, index_html)
	})
`, urlPrefix[:len(urlPrefix)-1]))

	}

	page.Code = codeBuilder.String()
	return StringFormatStructs(codeTpl, page), nil
}
