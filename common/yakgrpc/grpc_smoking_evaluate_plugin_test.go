package yakgrpc

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

func TestGRPCMUSTPASS_LANGUAGE_SMOKING_EVALUATE_PLUGIN(t *testing.T) {
	client, err := NewLocalClient()
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		code      string
		err       string
		codeTyp   string
		zeroScore bool // is score == 0 ?
	}
	TestSmokingEvaluatePlugin := func(tc testCase) {
		name, err := yakit.CreateTemporaryYakScript(tc.codeTyp, tc.code)
		if err != nil {
			t.Fatal(err)
		}
		rsp, err := client.SmokingEvaluatePlugin(context.Background(), &ypb.SmokingEvaluatePluginRequest{
			PluginName: name,
		})
		if err != nil {
			t.Fatal(err)
		}
		var checking = false
		fmt.Printf("result: %#v \n", rsp.String())
		if tc.zeroScore && rsp.Score != 0 {
			// want score == 0 but get !0
			t.Fatal("this test should have score = 0")
		}
		if !tc.zeroScore && rsp.Score == 0 {
			// want score != 0 but get 0
			t.Fatal("this test shouldn't have score = 0")
		}
		if tc.err == "" {
			if len(rsp.Results) != 0 {
				t.Fatal("this test should have no result")
			}
		} else {
			for _, r := range rsp.Results {
				if strings.Contains(r.String(), tc.err) {
					checking = true
				}
			}
			if !checking {
				t.Fatalf("should have %s", tc.err)
			}
		}
	}

	t.Run("test negative alarm", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `
yakit.AutoInitYakit()
handle = result => {
	yakit.Info("HELLO")
	risk.NewRisk("http://baidu.com", risk.cve(""))
}`,
			err:       "[Negative Alarm]",
			codeTyp:   "port-scan",
			zeroScore: false,
		})
	})

	t.Run("test undefine variable", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `
yakit.AutoInitYakit()
handle = result => {
	yakit.Info(bacd)
	risk.NewRisk("http://baidu.com", risk.cve(""))
}`,
			codeTyp:   "port-scan",
			err:       "Value undefine",
			zeroScore: true,
		})
	})

	t.Run("test just warning", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `
yakit.AutoInitYakit()
handle = result => {
}`,
			err:       "empty block",
			codeTyp:   "port-scan",
			zeroScore: false,
		})
	})

	t.Run("test yak ", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `
yakit.AutoInitYakit()

# Input your code!
			`,

			err:       "",
			codeTyp:   "yak",
			zeroScore: false,
		})

	})

	t.Run("test nuclei false positive", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `id: CVE-2024-32030

info:
  name: CVE-2024-32030 JMX Metrics Collection JNDI RCE
  severity: critical

requests:
  - method: GET
    path:
      - "{{BaseURL}}/api/clusters/malicious-cluster"
    matchers:
      - type: word
        part: body
        words:
          - "malicious-cluster"
`,
			err:       "误报",
			codeTyp:   "nuclei",
			zeroScore: false,
		})
	})

	t.Run("test nuclei positive", func(t *testing.T) {
		TestSmokingEvaluatePlugin(testCase{
			code: `id: WebFuzzer-Template-idZEfBnT
info:
  name: WebFuzzer Template idZEfBnT
  author: god
  severity: low
  description: write your description here
  reference:
  - https://github.com/
  - https://cve.mitre.org/
  metadata:
    max-request: 1
    shodan-query: ""
    verified: true
  yakit-info:
    sign: 39724ac438ac2b32ae79defc1f3eac22
variables:
  aa: '{{rand_char(5)}}'
  bb: '{{rand_char(6)}}'
http:
- method: POST
  path:
  - '{{RootURL}}/'
  headers:
    Content-Type: application/json
  body: "echo {{aa}}+{{bb}}"
  # attack: pitchfork
  max-redirects: 3
  matchers-condition: and
  matchers:
  - id: 1
    type: dsl
    part: body
    dsl:
    - '{{contains(body,aa+bb)}}'
    condition: and`,
			err:       "",
			codeTyp:   "nuclei",
			zeroScore: false,
		})
	})

}
