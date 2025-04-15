package messages

// client errors
const (
	ErrEncryption      = "encryption error"
	ErrDecryption      = "decryption error"
	ErrPageOut         = "failed to output the page"
	ErrBadRequest      = "failed to decode the request body"
	ErrSessionSet      = "failed to create a session"
	ErrNoToken         = "auth token is missing"
	ErrBadToken        = "invalid token"
	ErrNoSession       = "session could not be found"
	ErrNoCookie        = "cookie could not be found"
	ErrNoParams        = "incomplete request"
	ErrBadStudentID    = "invalid student ID"
	ErrBadTeacherID    = "invalid teacher ID"
	ErrBadTaskID       = "invalid task ID"
	ErrBadGrade        = "invalid grade"
	ErrNotTheirStudent = "you are not their student"
	ErrBadRating       = "invalid rating"
)

// logger errors
const (
	ErrKey             = "failed to get an encryption key"
	ErrHTML            = "failed to parse the html"
	ErrDecodeRequest   = "failed to decode the request body"
	ErrDecrypt         = "failed to decrypt data"
	ErrAuth            = "failed to authorize"
	ErrGenToken        = "failed to generate a token"
	ErrCeateAcc        = "failed to create an account"
	ErrParseToken      = "failed to parse JWT token"
	ErrDelSession      = "failed to delete session"
	ErrNoAuthToken     = "missing authToken cookie"
	ErrSessionNotFound = "session could not be found"
	ErrParseStudentID  = "invalid student UUID"
	ErrParseTeacherID  = "invalid teacher UUID"
	ErrParseTaskID     = "invalid task UUID"
	ErrParseGrade      = "invalid grade format"
	ErrParseRating     = "invalid rating format"
)

// repo errors
const (
	ErrNameTaken       = "username has been taken"
	ErrCred            = "username or password is incorrect"
	ErrUserNotFound    = "user could not be found"
	ErrTeacherNotFound = "teacher could no be found"
	ErrStudentNotFound = "student could not be found"
	ErrStudentReq      = "student request could not be found"
	ErrTeacherReq      = "teacher request could not be found"
	ErrTaskEmpty       = "task is empty"
	ErrSolutionEmpty   = "solution is empty"
	ErrTaskNotFound    = "task could not be found"
)

// misc
const (
	LogDetails         = "details"
	ServiceEncryption  = "encryption"
	ServiceAuth        = "auth"
	ServiceTasks       = "tasks"
	ServiceUsers       = "user"
	LogUserID          = "userID"
	LogSessionID       = "sessionID"
	LogUserRole        = "role"
	LogNeedRole        = "needed"
	LogReqPath         = "path"
	LogTaskID          = "taskID"
	LogGrade           = "grade"
	LogTotal           = "totalGrade"
	LogRating          = "rating"
	RoleTeacher        = "teacher"
	RoleStudent        = "student"
	ErrDecrEmpty       = "decryption data is empty"
	ErrDecrPaddingSize = "invalid padding size"
	ErrDecrPaddindByte = "invalid padding bytes"
	ErrDecrCipher      = "ciphertext is not a multiple of the block size"
)

// client status
const (
	StatusAuth         = "authorized"
	StatusLogOut       = "logged out"
	StatusNoPermission = "permission denied"
	StatusTaskCreated  = "task was created"
	StatusTaskUpdated  = "task was updated"
	StatusRated        = "rating was added"
	StatusReqSent      = "request was sent"
	StatusReqAccepted  = "request was accepted"
	StatusReqDenied    = "request was denied"
	StatusReqCanceled  = "request was canceled"
	StatusUpdated      = "profile updated"
)

// logger status
const (
	StatusUserAuth         = "user authorized"
	StatusUserLogOut       = "user logged out"
	StatusUserNoPermission = "permission denied"
	StatusUserTaskCreated  = "user created a task"
	StatusUserSolution     = "solution added"
	StatusUserGrade        = "grade added"
	StatusUserRated        = "rating added"
	StatusUserReqSent      = "request sent"
	StatusUserReqAccepted  = "request accepted"
	StatusUserReqDenied    = "request denied"
	StatusUserReqCanceled  = "request canceled"
	StatusUserUpdated      = "profile updated"
)

// request fields
const (
	ReqUsername   = "username"
	ReqPassword   = "password"
	ReqRole       = "role"
	ReqStudentID  = "studentID"
	ReqTaskName   = "taskName"
	ReqFileName   = "fileName"
	ReqTaskID     = "taskID"
	ReqGrade      = "grade"
	ReqOrderBy    = "orderBy"
	ReqOrderField = "orderField"
	ReqSpecialty  = "specialty"
	ReqTeacherID  = "teacherID"
	ReqRating     = "rating"
)

// cookie names
const (
	CookieAuthToken = "authToken"
	CookieUserRole  = "userRole"
)
