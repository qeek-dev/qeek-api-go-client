package qts_test

import (
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/codeskyblue/go-sh"
	"github.com/stretchr/testify/assert"

	"github.com/qeek-dev/qeek-api-go-client/qbus/qts/v1"
	"github.com/qeek-dev/qeek-api-go-client/test"
)

type QtsTestCaseSuite struct {
	qts *qts.Service

	validSidResponse, inValidSidResponse string
}

func setupSidTestCase(t *testing.T) (QtsTestCaseSuite, func(t *testing.T)) {
	s := QtsTestCaseSuite{}
	s.validSidResponse = `{"code":200,"errorCode":0,"errorMsg":"","result":null}`
	s.inValidSidResponse = `{"code":400,"errorCode":4000201,"errorMsg":"NAS sid is not valid","result":null}`

	s.qts = qts.NewClient("com.qnap.dj2", true)

	return s, func(t *testing.T) {

	}
}

func TestLoginCall_Do(t *testing.T) {
	s, teardownTestCase := setupSidTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name               string
		givenQbusNameSpace string
		givenUserName      string
		givenPassword      string

		//wantErr string
		wantErrCode qts.QtsErrCode

		setupSubTest test.SetupSubTest
	}{
		{
			name:          "success",
			givenUserName: "admin",
			givenPassword: "zxcv",
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(_ *sh.Session) (out []byte, err error) {
					return []byte(`{"code": 200,"errorCode": 0,"errorMsg": "","result": {"authPassed": 1,"authSid": "uyvoud8k","isAdmin": 1}}`), nil
				})
				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:          "fail with invalid password",
			givenUserName: "admin",
			givenPassword: "dddd",
			wantErrCode:   qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(_ *sh.Session) (out []byte, err error) {
					return []byte(`{"code": 400,"errorCode": 4000203,"errorMsg": "Authentication failed","result": null}`), nil
				})
				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:               "fail with qbus not found",
			givenQbusNameSpace: "com.qnap.dj2",
			givenUserName:      "admin",
			givenPassword:      "zxcv",
			wantErrCode:        qts.QtsErrorInternalError,
			setupSubTest:       test.EmptySubTest(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupSubTest(t)
			defer teardownSubTest(t)

			err := s.qts.Login().UserName(tc.givenUserName).Password(tc.givenPassword).Do()
			if err != nil {
				if c, ok := err.(*qts.QtsErr); ok {
					assert.Equal(t, c.Code, tc.wantErrCode, "An error was expected")
				} else {
					t.Fatalf("%v, unexpected error", err)
				}
			}
		})
	}
}

func TestVerifySidCall_Do(t *testing.T) {
	s, teardownTestCase := setupSidTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name     string
		givenSid string

		wantErrCode qts.QtsErrCode

		setupSubTest test.SetupSubTest
	}{
		{
			name:     "success with valid sid",
			givenSid: "hcm3ipzf",
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					return []byte(s.validSidResponse), nil
				})
				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:        "fail with invalid sid",
			givenSid:    "oh0n736f",
			wantErrCode: qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					return []byte(s.inValidSidResponse), nil
				})
				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:        "fail with empty sid",
			givenSid:    "",
			wantErrCode: qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					return []byte(`{"code": 400,"errorCode": 4000200,"errorMsg": "'sid' is not specified or not found.","result": null}`), nil
				})
				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:         "fail with qbus not found",
			givenSid:     "oh0n736f",
			wantErrCode:  qts.QtsErrorInternalError,
			setupSubTest: test.EmptySubTest(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupSubTest(t)
			defer teardownSubTest(t)

			err := s.qts.Verify().Sid(tc.givenSid).Do()
			if err != nil {
				if c, ok := err.(*qts.QtsErr); ok {
					assert.Equal(t, c.Code, tc.wantErrCode, "An error was expected")
				} else {
					t.Fatalf("%v, unexpected error", err)
				}
			}
		})
	}
}

func TestNasUsersCall_Do(t *testing.T) {
	s, teardownTestCase := setupSidTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name string

		wantNasAccount []qts.NasUserResult
		wantErrCode    qts.QtsErrCode

		setupSubTest test.SetupSubTest
	}{
		{
			name: "success with valid sid",
			wantNasAccount: []qts.NasUserResult{
				{"garychen@qnap.com", 1, []string{"administrators", "everyone"}, "auto", "admin", "/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg"},
				{"hykuan@qnap.com", 0, []string{"everyone"}, "TCH", "hykuan", "/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/hykuan/avatar/portrait.jpg"},
				{"cutedogspark@gmail.com", 1, []string{"administrators", "everyone"}, "auto", "gary", "/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/gary/avatar/portrait.jpg"},
			},
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.validSidResponse)
					case 1:
						out = []byte(`{"code":200,"errorCode":0,"errorMsg":"","result":[{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg","email":"garychen@qnap.com","enable":1,"group":["administrators","everyone"],"lang":"auto","name":"admin"},{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/hykuan/avatar/portrait.jpg","email":"hykuan@qnap.com","enable":0,"group":["everyone"],"lang":"TCH","name":"hykuan"},{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/gary/avatar/portrait.jpg","email":"cutedogspark@gmail.com","enable":1,"group":["administrators","everyone"],"lang":"auto","name":"gary"}]}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("valid-sid-mocked").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:        "fail with invalid sid",
			wantErrCode: qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.inValidSidResponse)
					case 1:
						out = []byte(`{"code":200,"errorCode":0,"errorMsg":"","result":[{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg","email":"garychen@qnap.com","enable":1,"group":["administrators","everyone"],"lang":"auto","name":"admin"},{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/hykuan/avatar/portrait.jpg","email":"hykuan@qnap.com","enable":0,"group":["everyone"],"lang":"TCH","name":"hykuan"},{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/gary/avatar/portrait.jpg","email":"cutedogspark@gmail.com","enable":1,"group":["administrators","everyone"],"lang":"auto","name":"gary"}]}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("invalid-sid-mocked").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:         "fail with qbus not found",
			wantErrCode:  qts.QtsErrorInternalError,
			setupSubTest: test.EmptySubTest(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupSubTest(t)
			defer teardownSubTest(t)

			na, err := s.qts.Users().Do()
			if err != nil {
				if c, ok := err.(*qts.QtsErr); ok {
					assert.Equal(t, c.Code, tc.wantErrCode, "An error was expected")
				} else {
					t.Fatalf("%v, unexpected error", err)
				}
			} else {
				for _, a := range tc.wantNasAccount {
					assert.Contains(t, na, a)
				}
			}
		})
	}
}

func TestNasUserCall_Do(t *testing.T) {
	s, teardownTestCase := setupSidTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name          string
		givenValidSid string
		givenUsername string

		wantNasAccount qts.NasUserResult
		wantErrCode    qts.QtsErrCode

		setupSubTest test.SetupSubTest
	}{
		{
			name:          "get nas account success",
			givenValidSid: "hcm3ipzf",
			givenUsername: "admin",
			wantNasAccount: qts.NasUserResult{
				"garychen@qnap.com", 1, []string{"administrators", "everyone"}, "auto", "admin", "/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg",
			},
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.validSidResponse)
					case 1:
						out = []byte(`{"code":200,"errorCode":0,"errorMsg":"","result":{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg","email":"garychen@qnap.com","enable":1,"group":["administrators","everyone"],"lang":"auto","name":"admin"}}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("valid-sid-mocked").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:          "fail with no match route for the path",
			givenValidSid: "oh0n736f",
			givenUsername: "ddd",
			wantErrCode:   qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.validSidResponse)
					case 1:
						out = []byte(`{"code":400,"errorCode":4000202,"errorMsg":"User dfdf not exist","result": null}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("valid-sid-mocked").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:          "fail with qbus not found",
			givenValidSid: "oh0n736f",
			wantErrCode:   qts.QtsErrorInternalError,
			setupSubTest:  test.EmptySubTest(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupSubTest(t)
			defer teardownSubTest(t)

			na, err := s.qts.User().UserName(tc.givenUsername).Do()
			if err != nil {
				if c, ok := err.(*qts.QtsErr); ok {
					assert.Equal(t, c.Code, tc.wantErrCode, "An error was expected")
				} else {
					t.Fatalf("%v, unexpected error", err)
				}
			} else {
				assert.EqualValues(t, na, tc.wantNasAccount)
			}
		})
	}
}

func TestNasMeCall_Do(t *testing.T) {
	s, teardownTestCase := setupSidTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name string

		wantNasMe   qts.NasUserResult
		wantErrCode qts.QtsErrCode

		setupSubTest test.SetupSubTest
	}{
		{
			name: "get nas me success",
			wantNasMe: qts.NasUserResult{
				"garychen@qnap.com", 1, []string{}, "auto", "admin", "/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg",
			},
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.validSidResponse)
					case 1:
						out = []byte(`{"code":200,"errorCode":0,"errorMsg":"","result":{"user":"admin"}}`)
					case 2:
						out = []byte(`{"code":200,"errorCode":0,"errorMsg":"","result":{"avatar":"/share/CACHEDEV1_DATA/.qpkg/DJ2-Live-X/middleware/qeek/../../tmp/share/user/nas/admin/avatar/portrait.jpg","email":"garychen@qnap.com","enable":1,"group":[],"lang":"auto","name":"admin"}}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("hcm3ipzf").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:        "get nas me fail with invalid sid",
			wantErrCode: qts.QtsErrorBadRequest,
			setupSubTest: func(t *testing.T) func(t *testing.T) {
				mockCount := 0
				monkey.PatchInstanceMethod(reflect.TypeOf((*sh.Session)(nil)), "Output", func(ss *sh.Session) (out []byte, err error) {
					switch mockCount {
					case 0:
						out = []byte(s.inValidSidResponse)
					case 1:
						out = []byte(`{"code":400,"errorCode":4000201,"errorMsg":"NAS sid is not valid","result":null}`)
					}
					mockCount++
					return
				})

				s.qts.Verify().Sid("invalid-sid").Do()

				return func(t *testing.T) {
					defer monkey.UnpatchAll()
				}
			},
		},
		{
			name:         "fail with qbus not found",
			wantErrCode:  qts.QtsErrorInternalError,
			setupSubTest: test.EmptySubTest(),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupSubTest(t)
			defer teardownSubTest(t)

			na, err := s.qts.Me().Do()
			if err != nil {
				if c, ok := err.(*qts.QtsErr); ok {
					assert.Equal(t, c.Code, tc.wantErrCode, "An error was expected")
				} else {
					t.Fatalf("%v, unexpected error", err)
				}
			} else {
				assert.EqualValues(t, na, tc.wantNasMe)
			}
		})
	}
}
