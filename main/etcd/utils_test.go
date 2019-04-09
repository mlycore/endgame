package etcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
)

func Test_ExecInPod(t *testing.T) {

}

func Test_AdmissionReviewEncoding(t *testing.T) {
	dataSet := []struct {
		ar       *v1beta1.AdmissionReview
		expected interface{}
		message  string
	}{
		{
			ar:      nil,
			message: "nil ar",
		},
		{
			ar:      &v1beta1.AdmissionReview{},
			message: "empty ar",
		},
	}

	for _, d := range dataSet {
		t.Logf("%s\n", string(admissionReviewEncoding(d.ar)))
		assert.Equal(t, d.expected, admissionReviewEncoding(d.ar), d.message)
	}
}
