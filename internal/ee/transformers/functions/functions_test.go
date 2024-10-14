package ee_transformer_fns

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/stretchr/testify/require"
)

// func Test_TransformPiiText(t *testing.T) {
// 	env := bloblang.NewEmptyEnvironment()
// 	mockanalyze := presidioapi.NewMockAnalyzeInterface(t)
// 	mockanon := presidioapi.NewMockAnonymizeInterface(t)
// 	err := NewBloblTransformPiiText(env, mockanalyze, mockanon, &mgmtv1alpha1.TransformPiiText{})
// 	require.NoError(t, err)

// 	mockanalyze.On("PostAnalyzeWithResponse", mock.Anything, mock.Anything).
// 		Return(&presidioapi.PostAnalyzeResponse{
// 			JSON200: &[]presidioapi.RecognizerResultWithAnaysisExplanation{
// 				{
// 					Start:      13,
// 					End:        21,
// 					EntityType: "name",
// 					Score:      100,
// 				},
// 			},
// 		}, nil)

// 	mockText := "my name is asdf and I am 100 years old"
// 	mockanon.On("PostAnonymizeWithResponse", mock.Anything, mock.Anything).
// 		Return(&presidioapi.PostAnonymizeResponse{
// 			JSON200: &presidioapi.AnonymizeResponse{Text: &mockText},
// 		}, nil)

// 	exec, err := env.Parse(`root = transform_pii_text(value:"my name is john doe and I am 100 years old")`)
// 	require.NoError(t, err)
// 	output, err := exec.Query(nil)
// 	require.NoError(t, err)

// 	value, ok := output.(string)
// 	require.True(t, ok)
// 	require.Equal(t, mockText, value)
// }

func Test_getDefaultAnonymizer(t *testing.T) {
	t.Run("redact", func(t *testing.T) {
		actual, err := getDefaultAnonymizer(&mgmtv1alpha1.PiiAnonymizer{
			Config: &mgmtv1alpha1.PiiAnonymizer_Redact_{
				Redact: &mgmtv1alpha1.PiiAnonymizer_Redact{},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	t.Run("replace", func(t *testing.T) {
		newval := "newval"
		actual, err := getDefaultAnonymizer(&mgmtv1alpha1.PiiAnonymizer{
			Config: &mgmtv1alpha1.PiiAnonymizer_Replace_{
				Replace: &mgmtv1alpha1.PiiAnonymizer_Replace{
					Value: &newval,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	t.Run("hash", func(t *testing.T) {
		sha256 := mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_SHA512
		actual, err := getDefaultAnonymizer(&mgmtv1alpha1.PiiAnonymizer{
			Config: &mgmtv1alpha1.PiiAnonymizer_Hash_{
				Hash: &mgmtv1alpha1.PiiAnonymizer_Hash{
					Algo: &sha256,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	t.Run("mask", func(t *testing.T) {
		maskingChar := "*"
		charsTomask := int32(5)
		fromend := false
		actual, err := getDefaultAnonymizer(&mgmtv1alpha1.PiiAnonymizer{
			Config: &mgmtv1alpha1.PiiAnonymizer_Mask_{
				Mask: &mgmtv1alpha1.PiiAnonymizer_Mask{
					MaskingChar: &maskingChar,
					CharsToMask: &charsTomask,
					FromEnd:     &fromend,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	t.Run("default", func(t *testing.T) {
		actual, err := getDefaultAnonymizer(nil)
		require.NoError(t, err)
		require.NotNil(t, actual)
	})
}

func Test_toPresidioHashType(t *testing.T) {
	t.Run("md5", func(t *testing.T) {
		actual := toPresidioHashType(mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_MD5)
		require.Equal(t, presidioapi.Md5, actual)
	})

	t.Run("sha256", func(t *testing.T) {
		actual := toPresidioHashType(mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_SHA256)
		require.Equal(t, presidioapi.Sha256, actual)
	})

	t.Run("sha512", func(t *testing.T) {
		actual := toPresidioHashType(mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_SHA512)
		require.Equal(t, presidioapi.Sha512, actual)
	})

	t.Run("default", func(t *testing.T) {
		actual := toPresidioHashType(mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_UNSPECIFIED)
		require.Equal(t, presidioapi.Md5, actual)
	})
}

func Test_handleAnonRespErr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := handleAnonRespErr(nil)
		require.Error(t, err)
	})

	t.Run("400", func(t *testing.T) {
		errtxt := "400 err"
		err := handleAnonRespErr(&presidioapi.PostAnonymizeResponse{
			JSON400: &presidioapi.N400BadRequest{Error: &errtxt},
		})
		require.Error(t, err)
	})

	t.Run("422", func(t *testing.T) {
		errtxt := "422 err"
		err := handleAnonRespErr(&presidioapi.PostAnonymizeResponse{
			JSON422: &presidioapi.N422UnprocessableEntity{Error: &errtxt},
		})
		require.Error(t, err)
	})

	t.Run("nil 200", func(t *testing.T) {
		err := handleAnonRespErr(&presidioapi.PostAnonymizeResponse{})
		require.Error(t, err)
	})

	t.Run("valid 200", func(t *testing.T) {
		err := handleAnonRespErr(&presidioapi.PostAnonymizeResponse{
			JSON200: &presidioapi.AnonymizeResponse{},
		})
		require.NoError(t, err)
	})
}