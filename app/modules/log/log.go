package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"web/config"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs" // ログローテーション
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

// ##########################################
// ログカテゴリー
// ##########################################
var (
	accessLog   *logrus.Logger // アクセスログ(ログレベル: INFO)
	errorLog    *logrus.Logger // エラーログ(ログレベル: ERROR, FATAL)
	securityLog *logrus.Logger // セキュリティログ(ログレベル: INFO, FATAL)
	debugLog    *logrus.Logger // デバッグログ(ログレベル: INFO)
)

type customStdFormatter struct {
	logrus.TextFormatter
}

func (f *customStdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = 36 // シアン
	case logrus.WarnLevel:
		levelColor = 33 // 黄
	case logrus.ErrorLevel:
		levelColor = 31 // 赤
	case logrus.FatalLevel, logrus.PanicLevel:
		levelColor = 41 // 赤反転
	default:
		levelColor = 34 // 青
	}
	// フォーマット: [INFO] [2021-10-06 15:49:23] message.....
	return []byte(fmt.Sprintf("\x1b[%dm[%s]\x1b[0m [%s] %s", levelColor, strings.ToUpper(entry.Level.String()), entry.Time.Format(f.TimestampFormat), entry.Message)), nil
}

// パッケージ初期化
func init() {
	_initAllLogger()
}

func _initAllLogger() {
	// アクセスログ初期化
	if config.GetEnv().AccessLog {
		// #############################################################
		// アクセスログ
		// 	・デバッグ時はローテーションファイルと標準出力にログを出力する
		// 	・リリース時はローテーションファイルのみにログを出力する
		//	・GinのログとログレベルInfoのログを出力する
		// #############################################################
		// ローテーションファイル出力作成
		writer := _initRotateLogger(config.GetEnv().AccessLogPath)

		accessLog = logrus.New() // 標準出力用
		accessLog.SetOutput(os.Stdout)
		accessLog.SetFormatter(&customStdFormatter{ // 標準出力用のログフォーマット
			logrus.TextFormatter{
				ForceColors:     true,
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			}})
		accessLog.Hooks.Add(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.FatalLevel: writer,
				logrus.PanicLevel: writer,
				logrus.DebugLevel: writer,
				logrus.TraceLevel: writer,
			},
			&easy.Formatter{ // ファイル出力用のログフォーマット
				TimestampFormat: "2006-01-02 15:04:05",
				LogFormat:       "[%lvl%] [%time%] %msg%",
			},
		))

		// Ginのログはファイルと標準出力にログを出力
		if gin.IsDebugging() {
			gin.DefaultWriter = io.MultiWriter(writer, os.Stdout)
		} else { // リリース時はファイルにのみ出力
			gin.DefaultWriter = writer
		}
	}

	// エラーログ初期化
	if config.GetEnv().ErrorLog {
		writer := _initLogger(config.GetEnv().ErrorLogPath)
		errorLog = logrus.New() // 標準エラー出力用
		errorLog.SetOutput(os.Stderr)
		errorLog.SetFormatter(&customStdFormatter{ // 標準エラー出力用のログフォーマット
			logrus.TextFormatter{
				ForceColors:     true,
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			}})
		errorLog.Hooks.Add(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.FatalLevel: writer,
				logrus.PanicLevel: writer,
				logrus.DebugLevel: writer,
				logrus.TraceLevel: writer,
			},
			&easy.Formatter{ // ファイル出力用のログフォーマット
				TimestampFormat: "2006-01-02 15:04:05",
				LogFormat:       "[%lvl%] [%time%] %msg%",
			},
		))
	}

	// セキュリティログ初期化
	if config.GetEnv().SecurityLog {
		writer := _initLogger(config.GetEnv().SecurityLogPath)
		securityLog = logrus.New() // 標準出力用
		securityLog.SetOutput(os.Stdout)
		securityLog.SetFormatter(&customStdFormatter{ // 標準エラー出力用のログフォーマット
			logrus.TextFormatter{
				ForceColors:     true,
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			}})
		securityLog.Hooks.Add(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.FatalLevel: writer,
				logrus.PanicLevel: writer,
				logrus.DebugLevel: writer,
				logrus.TraceLevel: writer,
			},
			&easy.Formatter{ // ファイル出力用のログフォーマット
				TimestampFormat: "2006-01-02 15:04:05",
				LogFormat:       "[%lvl%] [%time%] %msg%",
			},
		))
	}

	// デバッグログ初期化
	if gin.IsDebugging() && config.GetEnv().DebugLog {
		writer := _initLogger(config.GetEnv().DebugLogPath)
		debugLog = logrus.New() // 標準エラー出力用
		debugLog.SetOutput(os.Stdout)
		debugLog.SetLevel(logrus.DebugLevel)       // *** デバッグレベル以上を出力 ***
		debugLog.SetFormatter(&customStdFormatter{ // 標準出力用のログフォーマット
			logrus.TextFormatter{
				ForceColors:     true,
				TimestampFormat: "2006-01-02 15:04:05",
				FullTimestamp:   true,
			}})
		debugLog.Hooks.Add(lfshook.NewHook(
			lfshook.WriterMap{
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.FatalLevel: writer,
				logrus.PanicLevel: writer,
				logrus.DebugLevel: writer,
				logrus.TraceLevel: writer,
			},
			&easy.Formatter{ // ファイル出力用のログフォーマット
				TimestampFormat: "2006-01-02 15:04:05",
				LogFormat:       "[%lvl%] [%time%] %msg%",
			},
		))
	}
}

// ログファイル作成
func _initLogger(path string) io.Writer {
	writer, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	return writer
}

// ローテーションを行うログファイルを作成
func _initRotateLogger(path string) io.Writer {
	writer, err := rotatelogs.New(
		path+"-%Y%m%d",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithRotationTime(time.Hour*24), // 1日ごとにローテーション
		rotatelogs.WithRotationCount(5),           //Keep only the closest N log files. select WithMaxAge() or WithRotationCount().
	)
	if err != nil {
		logrus.Errorf("config local file system for logger error: %v", err)
	}
	return writer
}

// #####################################
// アプリケーションイベントログ出力用(Info)
// #####################################
func Info(args ...interface{}) {
	if accessLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		accessLog.Info(s)
	}
}
func Infoln(args ...interface{}) {
	if accessLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		accessLog.Infoln(s)
	}
}
func Infof(format string, args ...interface{}) {
	if accessLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		accessLog.Infof(format, args...)
	}
}

// #####################################
// アプリケーションエラー用(Error, Fatal)
// #####################################
func ErrorStack(err error) {
	if errorLog != nil {
		err = errors.WithStack(err)
		msg := fmt.Sprintf("%+v", err)
		Error(msg)
	}
}
func Error(args ...interface{}) {
	if errorLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		errorLog.Error(s)
	}
}
func Errorln(args ...interface{}) {
	if errorLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		errorLog.Errorln(s)
	}
}
func Errorf(format string, args ...interface{}) {
	if errorLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		errorLog.Errorf(format, args...)
	}
}
func FatalStack(err error) {
	if errorLog != nil {
		err = errors.WithStack(err)
		msg := fmt.Sprintf("%+v", err)
		Fatal(msg)
	}
}
func Fatal(args ...interface{}) {
	if errorLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		errorLog.Fatal(s)
	}
}
func Fatalln(args ...interface{}) {
	if errorLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		errorLog.Fatalln(s)
	}
}
func Fatalf(format string, args ...interface{}) {
	if errorLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		errorLog.Fatalf(format, args...)
	}
}

// #####################################
// セキュリティ情報出力用(Security, SecurityAlert)
// #####################################
func Security(args ...interface{}) {
	if securityLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		securityLog.Info(s)
	}
}
func Securityln(args ...interface{}) {
	if securityLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		securityLog.Infoln(s)
	}
}
func Securityf(format string, args ...interface{}) {
	if securityLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		securityLog.Infof(format, args...)
	}
}

func SecurityAlert(args ...interface{}) {
	if securityLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		securityLog.Fatal(s)
	}
}
func SecurityAlertln(args ...interface{}) {
	if securityLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		securityLog.Fatalln(s)
	}
}
func SecurityAlertf(format string, args ...interface{}) {
	if securityLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		securityLog.Fatalf(format, args...)
	}
}

// #####################################
// デバッグ出力用(Debug, Print)
// デバッグ時: ファイル出力と標準出力
// デバッグオフ時: 出力なし
// ####################################
func Debug(args ...interface{}) {
	if gin.IsDebugging() && debugLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprint(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		debugLog.Debug(s)
	}
}
func Debugln(args ...interface{}) {
	if gin.IsDebugging() && debugLog != nil {
		// 一旦stringに変換して改行を追加
		s := fmt.Sprintln(args...)
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		debugLog.Debugln(s)
	}
}
func Debugf(format string, args ...interface{}) {
	if gin.IsDebugging() && debugLog != nil {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		debugLog.Debugf(format, args...)
	}
}

func Print(args ...interface{}) {
	Debug(args...)
}
func Println(args ...interface{}) {
	Debugln(args...)
}
func Printf(format string, args ...interface{}) {
	Debugf(format, args...)
}
