package proxy

import (
	"encoding/xml"
)

type XmlProxyResponse struct {
	XMLName xml.Name `xml:"http://www.yale.edu/tp/cas serviceResponse"`

	Failure *XmlProxyFailure
	Success *XmlProxySuccess
}

type XmlProxyFailure struct {
	XMLName xml.Name `xml:"proxyFailure"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

type XmlProxySuccess struct {
	XMLName     xml.Name `xml:"proxySuccess"`
	ProxyTicket string   `xml:"proxyTicket"`
}
