package proxy

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXmlProxyResponseSuccess(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:proxySuccess>
    <cas:proxyTicket>PT-XXXXXXXXXX</cas:proxyTicket>
  </cas:proxySuccess>
</cas:serviceResponse>
`

	var response XmlProxyResponse
	err := xml.Unmarshal([]byte(xmlData), &response)

	assert.Nil(t, err, "Expected err to be nil")
	assert.Nil(t, response.Failure, "Expected Success to be nil")
	assert.Equal(t, "PT-XXXXXXXXXX", response.Success.ProxyTicket, "Expected ProxyTicket to be 'PT-XXXXXXXXXX'")
}

func TestXmlProxyResponseFailure(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
	<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
		<cas:proxyFailure code="INVALID_TICKET">
			Invalid proxy ticket
		</cas:proxyFailure>
	</cas:serviceResponse>`

	var response XmlProxyResponse
	err := xml.Unmarshal([]byte(xmlData), &response)
	assert.Nil(t, err, "Expected err to be nil")
	assert.NotNil(t, response.Failure, "Expected Failure to be non-nil")
	assert.Equal(t, "INVALID_TICKET", response.Failure.Code, "Expected Code to be 'INVALID_TICKET'")
	assert.Equal(t, "Invalid proxy ticket", strings.TrimSpace(response.Failure.Message), "Expected Message to be 'Invalid proxy ticket'")
}
