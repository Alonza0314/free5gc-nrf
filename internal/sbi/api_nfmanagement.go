/*
 * NRF NFManagement Service
 *
 * NRF NFManagement Service
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package sbi

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"

	"github.com/free5gc/nrf/internal/logger"
	"github.com/free5gc/nrf/internal/util"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
)

func (s *Server) getNfRegisterRoute() []Route {
	// Since OAuth now have to use NFProfile to issue token, so we have to let NF to register without token
	return []Route{
		{
			"RegisterNFInstance",
			http.MethodPut,
			"/nf-instances/:nfInstanceID",
			s.HTTPRegisterNFInstance,
		},
	}
}

func (s *Server) getNfManagementRoute() []Route {
	return []Route{
		{
			"Index",
			http.MethodGet,
			"/",
			func(c *gin.Context) {
				c.JSON(http.StatusOK, "free5gc")
			},
		},
		{
			"DeregisterNFInstance",
			http.MethodDelete,
			"/nf-instances/:nfInstanceID",
			s.HTTPDeregisterNFInstance,
		},
		{
			"GetNFInstance",
			http.MethodGet,
			"/nf-instances/:nfInstanceID",
			s.HTTPGetNFInstance,
		},
		// Have another router group without Middlerware OAuth Check
		// {
		// 	"RegisterNFInstance",
		// 	http.MethodPut,
		// 	"/nf-instances/:nfInstanceID",
		// 	s.HTTPRegisterNFInstance,
		// },
		{
			"UpdateNFInstance",
			http.MethodPatch,
			"/nf-instances/:nfInstanceID",
			s.HTTPUpdateNFInstance,
		},
		{
			"GetNFInstances",
			http.MethodGet,
			"/nf-instances",
			s.HTTPGetNFInstances,
		},
		{
			"RemoveSubscription",
			http.MethodDelete,
			"/subscriptions/:subscriptionID",
			s.HTTPRemoveSubscription,
		},
		{
			"UpdateSubscription",
			http.MethodPatch,
			"/subscriptions/:subscriptionID",
			s.HTTPUpdateSubscription,
		},
		{
			"CreateSubscription",
			http.MethodPost,
			"/subscriptions",
			s.HTTPCreateSubscription,
		},
	}
}

// DeregisterNFInstance - Deregisters a given NF Instance
func (s *Server) HTTPDeregisterNFInstance(c *gin.Context) {
	nfInstanceID := c.Params.ByName("nfInstanceID")
	if nfInstanceID == "" {
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: "",
		}
		util.GinProblemJson(c, problemDetails)
		return
	}
	s.Processor().HandleNFDeregisterRequest(c, nfInstanceID)
}

// GetNFInstance - Read the profile of a given NF Instance
func (s *Server) HTTPGetNFInstance(c *gin.Context) {
	nfInstanceID := c.Params.ByName("nfInstanceID")
	if nfInstanceID == "" {
		problemDetails := &models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: "",
		}
		util.GinProblemJson(c, problemDetails)
		return
	}
	s.Processor().HandleGetNFInstanceRequest(c, nfInstanceID)
}

// RegisterNFInstance - Register a new NF Instance
func (s *Server) HTTPRegisterNFInstance(c *gin.Context) {
	// // step 1: retrieve http request body
	var nfprofile models.NrfNfManagementNfProfile

	requestBody, err := c.GetRawData()
	if err != nil {
		problemDetail := &models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		logger.NfmLog.Errorf("Get Request Body error: %+v", err)
		util.GinProblemJson(c, problemDetail)
		return
	}

	// step 2: convert requestBody to openapi models
	err = openapi.Deserialize(&nfprofile, requestBody, "application/json")
	if err != nil {
		details := "[Request Body] " + err.Error()
		pd := &models.ProblemDetails{
			Title:  "Malformed request syntax",
			Status: http.StatusBadRequest,
			Detail: details,
		}
		logger.NfmLog.Errorln(details)
		util.GinProblemJson(c, pd)
		return
	}

	s.Processor().HandleNFRegisterRequest(c, &nfprofile)
}

// UpdateNFInstance - Update NF Instance profile
func (s *Server) HTTPUpdateNFInstance(c *gin.Context) {
	// step 1: retrieve http request body
	requestBody, err := c.GetRawData()
	if err != nil {
		problemDetail := models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		logger.NfmLog.Errorf("Get Request Body error: %+v", err)
		c.JSON(http.StatusInternalServerError, problemDetail)
		return
	}
	nfInstanceID := c.Params.ByName("nfInstanceID")
	if nfInstanceID == "" {
		problemDetail := &models.ProblemDetails{
			Title:  "nfInstanceID Empty",
			Status: http.StatusBadRequest,
			Detail: "nfInstanceID not exist in request",
		}
		util.GinProblemJson(c, problemDetail)
		return
	}
	s.Processor().HandleUpdateNFInstanceRequest(c, requestBody, nfInstanceID)
}

// GetNFInstances - Retrieves a collection of NF Instances
func (s *Server) HTTPGetNFInstances(c *gin.Context) {
	nfType := c.Query("nf-type")
	limitParam := c.Query("limit")

	if nfType == "" || limitParam == "" {
		problemDetail := &models.ProblemDetails{
			Title:  "nfType or limitParam empty",
			Status: http.StatusBadRequest,
			Detail: fmt.Sprintf("nfType: %v, limitParam: %v", nfType, limitParam),
		}
		util.GinProblemJson(c, problemDetail)
		return
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		logger.NfmLog.Errorln("Error in string conversion: ", limit)
		problemDetails := &models.ProblemDetails{
			Title:  "Invalid Parameter",
			Status: http.StatusBadRequest,
			Detail: err.Error(),
		}
		util.GinProblemJson(c, problemDetails)
		return
	}
	if limit < 1 {
		problemDetails := &models.ProblemDetails{
			Title:  "Invalid Parameter",
			Status: http.StatusBadRequest,
			Detail: "limit must be greater than 0",
		}
		util.GinProblemJson(c, problemDetails)
		return
	}

	s.Processor().HandleGetNFInstancesRequest(c, nfType, limit)
}

// RemoveSubscription - Deletes a subscription
func (s *Server) HTTPRemoveSubscription(c *gin.Context) {
	subscriptionID := c.Params.ByName("subscriptionID")
	if subscriptionID == "" {
		problemDetail := &models.ProblemDetails{
			Title:  "subscriptionID Empty",
			Status: http.StatusBadRequest,
			Detail: "subscriptionID not exist in request",
		}
		util.GinProblemJson(c, problemDetail)
		return
	}
	s.Processor().HandleRemoveSubscriptionRequest(c, subscriptionID)
}

// UpdateSubscription - Updates a subscription
func (s *Server) HTTPUpdateSubscription(c *gin.Context) {
	requestBody, err := c.GetRawData()
	if err != nil {
		problemDetail := &models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		logger.NfmLog.Errorf("Get Request Body error: %+v", err)
		util.GinProblemJson(c, problemDetail)
		return
	}

	subscriptionID := c.Params.ByName("subscriptionID")
	if subscriptionID == "" {
		problemDetail := &models.ProblemDetails{
			Title:  "subscriptionID Empty",
			Status: http.StatusInternalServerError,
			Detail: "subscriptionID not exist in request",
		}
		util.GinProblemJson(c, problemDetail)
		return
	}

	s.Processor().HandleUpdateSubscriptionRequest(c, subscriptionID, requestBody)
}

// CreateSubscription - Create a new subscription
func (s *Server) HTTPCreateSubscription(c *gin.Context) {
	var subscription models.NrfNfManagementSubscriptionData

	// step 1: retrieve http request body
	requestBody, err := c.GetRawData()
	if err != nil {
		problemDetail := models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		logger.NfmLog.Errorf("Get Request Body error: %+v", err)
		c.JSON(http.StatusInternalServerError, problemDetail)
		return
	}

	// step 2: convert requestBody to openapi models
	err = openapi.Deserialize(&subscription, requestBody, "application/json")
	if err != nil {
		problemDetail := "[Request Body] " + err.Error()
		rsp := models.ProblemDetails{
			Title:  "Malformed request syntax",
			Status: http.StatusBadRequest,
			Detail: problemDetail,
		}
		logger.NfmLog.Errorln(problemDetail)
		c.JSON(http.StatusBadRequest, rsp)
		return
	}
	s.Processor().HandleCreateSubscriptionRequest(c, subscription)
}

// DecodeNfProfile - Only support []map[string]interface to []models.NfProfile
func (s *Server) DecodeNfProfile(source interface{}, format string) (models.NrfNfManagementNfProfile, error) {
	var target models.NrfNfManagementNfProfile

	// config mapstruct
	stringToDateTimeHook := func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if t == reflect.TypeOf(time.Time{}) && f == reflect.TypeOf("") {
			return time.Parse(format, data.(string))
		}
		return data, nil
	}

	config := mapstructure.DecoderConfig{
		DecodeHook: stringToDateTimeHook,
		Result:     &target,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return target, err
	}

	// Decode result to NfProfile structure
	err = decoder.Decode(source)
	if err != nil {
		return target, err
	}
	return target, nil
}
