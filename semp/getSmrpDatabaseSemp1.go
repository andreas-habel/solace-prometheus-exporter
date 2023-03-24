package semp

import (
	"encoding/xml"
	"errors"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Get system health information
func (e *Semp) GetSmrpDatabaseSemp1(ch chan<- prometheus.Metric) (ok float64, err error) {
	type Data struct {
		RPC struct {
			Show struct {
				Smrp struct {
					Database struct {
						Routers struct {
							Router []struct {
								Name      string `xml:"router-name"`
								NumBlocks int    `xml:"num-blocks"`
							} `xml:"router"`
						} `xml:"routers"`
					} `xml:"database"`
				} `xml:"smrp"`
			} `xml:"show"`
		} `xml:"rpc"`
		ExecuteResult struct {
			Result string `xml:"code,attr"`
		} `xml:"execute-result"`
	}

	command := "<rpc><show><smrp><database/></smrp></show></rpc>"
	body, err := e.postHTTP(e.brokerURI+"/SEMP", "application/xml", command)
	if err != nil {
		_ = level.Error(e.logger).Log("msg", "Can't scrape GetSmrpDatabaseSemp1", "err", err, "broker", e.brokerURI)
		return 0, err
	}
	defer body.Close()
	decoder := xml.NewDecoder(body)
	var target Data
	err = decoder.Decode(&target)
	if err != nil {
		_ = level.Error(e.logger).Log("msg", "Can't decode Xml GetSmrpDatabaseSemp1", "err", err, "broker", e.brokerURI)
		return 0, err
	}
	if target.ExecuteResult.Result != "ok" {
		_ = level.Error(e.logger).Log("msg", "unexpected result", "command", command, "result", target.ExecuteResult.Result, "broker", e.brokerURI)
		return 0, errors.New("unexpected result: see log")
	}

	for _, router := range target.RPC.Show.Smrp.Database.Routers.Router {
		ch <- prometheus.MustNewConstMetric(MetricDesc["SmrpDatabase"]["router_blocks"], prometheus.GaugeValue, float64(router.NumBlocks), router.Name)
	}

	return 1, nil
}
