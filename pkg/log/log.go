package log

// imports //////////////////////////////////////

import(
  "os"
  "regexp"
  "errors"

  "github.com/rs/zerolog"

  "github.com/callowaylc/mq/pkg"
)

// constants ////////////////////////////////////

// functions ////////////////////////////////////

func Init() { }

func Logger(trace string) zerolog.Logger {
  // determine reexp
  r := regexp.MustCompile(`^(.+?)#(.+?)\.(.+?)@(.+)$`)
  matches := r.FindStringSubmatch(trace)
  public, _ := regexp.MatchString("^[A-Z]", matches[3])

  // return json logger unless the env LOGGEr=true exists
  // at the time of call
  v, ok := os.LookupEnv("LOGGER"); if ok && v == "true" {
    return zerolog.
      New(os.Stderr).
      With().
      Caller().
      Str("Trace", trace).
      Str("Project", matches[1]).
      Str("Package", matches[2]).
      Str("Function", matches[3]).
      Str("File", matches[4]).
      Bool("Public", public).
      Logger()
  }

  // otherwise return a nop logger
  return zerolog.Nop()
}

func ParseLevel(l string) (zerolog.Level, error) {
  // attempt to parse level, using the built-in parser,
  // and then falling back to regex match; an error is
  // return if a level can't be determined
  logger := Logger(pkg.Trace("ParseLevel", "log"))
  logger.Info().
    Str("level", l).
    Msg("Enter")
  defer logger.Info().Msg("Exit")

  level := zerolog.InfoLevel
  level, err := zerolog.ParseLevel(l)

  if err != nil {
    switch {
    case match(`(?i)debug`, l):
      level = zerolog.DebugLevel
    case match(`(?i)notice`, l):
      level = zerolog.InfoLevel
    case match(`(?i)warn`, l):
      level = zerolog.WarnLevel
    case match(`(?i)err`, l):
      level = zerolog.ErrorLevel
    case match(`(?i)(crit|alert)`, l):
      level = zerolog.FatalLevel
    case match(`(?i)emerg`, l):
      level = zerolog.PanicLevel

    default:
      return 0, errors.New("Failed to determine level")
    }
  }

  return level, nil
}

func match(pattern, subject string) bool {
  if ok, _ := regexp.MatchString(pattern, subject); ok {
    return true
  }

  return false
}
