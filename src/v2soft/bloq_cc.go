/**
*  Note:
*    1. Stub is passed as a parameter in every function.
*		  This function is used to access and modify ledgers.
*    2. All the functions are called by their respective NodeJS functions.
*    3. None of the functions can be called explicitly.
 */

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	//Blockchain Crypto Service Provider
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"

	//Client Identity Chaincode Library
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("bloq_cc.go")

type Budget struct {
	Fixed_Cost    float64 `json:"fixed_cost"`
	Variable_Cost float64 `json:"variable_cost"`
}

type MilestoneCosts struct {
	Cost    float64 `json:"cost"`
	Comment string  `json:"comment"`
}

type SiteFixed struct {
	CourierCosts   float64        `json:"courier_costs"`
	IRB_fees       float64        `json:"irb_fees"`
	MilestoneCost  MilestoneCosts `json:"milestone_cost"`
	TotalCost      float64        `json:"total_cost"`
	RegulatoryFees float64        `json:"regulatory_fees"`
	ManpowerCosts  float64        `json:"manpower_costs"`
}

type SiteVariable struct {
	CourierCosts      float64        `json:"courier_costs"`
	MilestoneCost     MilestoneCosts `json:"milestone_cost"`
	TotalCost         float64        `json:"total_cost"`
	AESAEUPCost       float64        `json:"ae_sae_up_cost"`
	PatientVisits     float64        `json:"patient_visits"`
	Diagnostics       float64        `json:"diagnostics"`
	TravelCostpatient float64        `json:"travel_cost_patient"`
	Imaging           float64        `json:"imaging"`
}

type SiteBudgetDetails struct {
	Fixed_Cost     float64      `json:"fixed_cost"`
	Variable_Cost  float64      `json:"variable_cost"`
	FixedBudget    SiteFixed    `json:"fixed_budget"`
	VariableBudget SiteVariable `json:"variable_budget"`
}

type TrialBudget struct {
	Budget       Budget        `json:"budget"`
	Expenditure  Budget        `json:"expenditure"`
	Transactions []Transaction `json:"transactions"`
}

type ChargeBreakUp struct {
	StripeCharges   string `json:"stripe_charges"`
	PlatformCharges string `json:"platform_charges"`
	CreditedAmount  string `json:"credited_amount"`
}

type Transaction struct {
	ChargeID            string        `json:"charge_id"`
	AmountCharged       float64       `json:"amount_charged"`
	Purpose             string        `json:"purpose"`
	DebitCustomerID     string        `json:"debit_customer_id"`
	CreditCustomerID    string        `json:"credit_customer_id"`
	SourceAccountName   string        `json:"source_account_name"`
	SourceBankName      string        `json:"source_bank_name"`
	SourceAccountLast4  string        `json:"source_account_last4"`
	RecipientName       string        `json:"recipient_name"`
	ReceipientAccountID string        `json:"recipient_account_id"`
	TimeStamp           string        `json:"timestamp"`
	ChargeBreakUp       ChargeBreakUp `json:"charge_breakup"`
}

type SiteBudget struct {
	Site_ID      float64           `json:"site_id"`
	Budget       SiteBudgetDetails `json:"budget"`
	Expenditure  Budget            `json:"expenditure"`
	Transactions []Transaction     `json:"transactions"`
}

type TotalBudget struct {
	Protocol_ID string       `json:"protocol_id"`
	TrialBudget TrialBudget  `json:"trial_budget"`
	SiteBudget  []SiteBudget `json:"site_budget"`
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
	bccspInst bccsp.BCCSP
}

//Init Exported functoin
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	var A string // Entities
	var Aval int // Asset holdings
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//Invoke -
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("=== Invoke ===")
	function, args := stub.GetFunctionAndParameters()

	if function == "create_trial" {
		return t.createTrial(stub, args)
	} else if function == "query_trial" {
		return t.queryTrial(stub, args)
	} else if function == "create_site_sel_que" {
		return t.createSiteSelectionQue(stub, args)
	} else if function == "query_site_sel_que" {
		return t.querySiteSelectionQue(stub, args)
	} else if function == "create_pre_site_visit" {
		return t.createPreSiteVisit(stub, args)
	} else if function == "query_pre_site_visit" {
		return t.queryPreSiteVisit(stub, args)
	} else if function == "create_site_selection_visit" {
		return t.createSiteSelectionVisit(stub, args)
	} else if function == "query_site_selection_visit" {
		return t.querySiteSelectionVisit(stub, args)
	} else if function == "query_budget" {
		return t.queryBudget(stub, args)
	} else if function == "create_site_initiate" {
		return t.createSiteInitiate(stub, args)
	} else if function == "query_site_initiate" {
		return t.querySiteInitiate(stub, args)
	} else if function == "update_budget" {
		return t.updateBudget(stub, args)
	} else if function == "create_subject" {
		return t.createSubject(stub, args)
	} else if function == "query_subject" {
		return t.querySubject(stub, args)
	} else if function == "create_crf" {
		return t.createCrf(stub, args)
	} else if function == "query_crf" {
		return t.queryCrf(stub, args)
	} else if function == "get_history_for_crf" {
		return t.getHistoryForCrf(stub, args)
	} else if function == "create_protocol_deviation" {
		return t.createProtocolDeviation(stub, args)
	} else if function == "query_protocol_deviation" {
		return t.queryProtocolDeviation(stub, args)
	}

	return shim.Error("Invalid invoke function name: " + function)
}

/**
 * @Description
 *	- Accepts the trial id and the trial details and creates a new block storing the trial
 *		details against that particular trial id.
 *	- Creates budget allocated in the trial budget details.
 *
 * @Caller
 *	- Create_protocol
 *
 * @params
 *	- []string: args
 *		- args is a string array containing all the required details for creating a new trial.
 *			- args[0]: trial id for the new trial.
 *			- args[1]: details for the new trial.
 *				protocol_id:
 *         		  type: string
 *         		  required: true
 *       		protocol:
 *       		  type: object
 *       		  required: true
 *       		  properties:
 *       		    protocol_details:
 *       		      type: object
 *       		      properties:
 *       		        title:
 *       		          type: string
 *       		          required: true
 *       		        number_of_sub:
 *       		          type: integer
 *       		        therapeutic_agents:
 *       		          type: array
 *       		          items:
 *       		            type: object
 *       		            properties:
 *       		              name:
 *       		                type: string
 *       		              class:
 *       		                type: string
 *       		    documents_by_sponsor:
 *       		      type: array
 *       		      required: true
 *       		      items:
 *       		        type: object
 *       		        properties:
 *       		          title:
 *       		            type: string
 *       		          type:
 *       		            type: string
 *       		          document_url:
 *       		            type: string
 *       		          version:
 *       		            type: string
 *       		          timestamp:
 *       		            type: string
 *       		            format: date-time
 *       		    study_details:
 *       		      type: object
 *       		      properties:
 *       		        study_type:
 *       		          type: string
 *       		        number_of_treatments:
 *       		          type: integer
 *       		        total_treatment_duration:
 *       		          type: integer
 *       		        test_article:
 *       		          type: array
 *       		          items:
 *       		            type: object
 *       		            properties:
 *       		              name_of_drug:
 *       		                type: string
 *       		              class_of_drug:
 *       		                type: string
 *       		              dosage:
 *       		                type: number
 *       		                format: float
 *       		              route_of_administration:
 *       		                type: string
 *       		              frequency_of_administration:
 *       		                type: string
 *
 * @returns
 *	- payload: nil
 *		- returns nil if the trial is successfully created.
 *
 * @returns
 *	- payload: string
 *		- returns error if the trial creation was unsuccessful.
 *			- Incorrect number of arguements passed: "Incorrect number of arguments.
 *				Expecting Key-Value pair"
 *			- Budget creation failed: "Failed to generate event"
 *			- Trial creation failed: "Failed to create protocol"
 *
 * @changelog
 */

func (t *SimpleChaincode) createTrial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Trial ===")

	// Check the number of arguements required for trial creation
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Key-Value pair")
	}

	// Can be used later to have the role validation in the blockchain
	_, _, _ = cid.GetAttributeValue(stub, "role")

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	// Entities
	protocolID := args[0]
	protocolDetails := args[1]

	// Creates budget with the trial budget details
	// _, err = t.createBudget(stub, transientMap, protocolID, protocolDetails)
	// if err != nil {
	// 	return shim.Error("Failed to generate event")
	// }

	// Encrypts and stores the trial data in a new block
	err = encryptAndPutState(stub, t.bccspInst, transientMap, protocolID, []byte(protocolDetails))
	if err != nil {
		return shim.Error("Failed to create protocol")
	}

	return shim.Success(nil)
}

/**
 * @Description
 *	- Creates budget according to that particular trial data.
 *
 * @params
 *	- []byte: transientMap
 *		- transientMap is a byte array containing the encryption key and IV to encrypt
 *			or decrypt the data	going into the blockchain or getting out of the blockchain.
 *	- string: protocolID
 *		- id of the trial for which the budget is to be created.
 *	- string: protocolDetails
 *		- trial details containing the budget details required for creating the budget.
 *
 * @returns
 *	- payload: true, nil
 *		- returns true, nil if the budget is successfully created.
 *
 * @returns
 *	- payload: bool, string
 *		- returns false, error if the budget creation was unsuccessful.
 *			- Budget creation failed: "Failed to create budget".
 *
 * @changelog
 */

func (t *SimpleChaincode) createBudget(stub shim.ChaincodeStubInterface, transientMap map[string][]byte, protocolID string, protocolDetails string) (bool, error) {
	logger.Info("=== CreateBudget ===")

	// Entities
	var protocolJSON interface{}

	// Fetching the budget JSON for getting fixed and variable costs
	json.Unmarshal([]byte(protocolDetails), &protocolJSON)

	// Entities
	protocolMap := protocolJSON.(map[string]interface{})
	budgetMap := protocolMap["protocol"].(map[string]interface{})

	budgets := budgetMap["budget"].(map[string]interface{})

	fixedCosts := budgets["fixed_cost"].(map[string]interface{})
	variableCosts := budgets["variable_cost"].(map[string]interface{})

	// Intializing the trial budget with fixed and variable costs
	trialBudget := Budget{fixedCosts["total_cost"].(float64), variableCosts["total_cost"].(float64)}

	// Initializing the trial expenditure with fixed and variable costs
	trialExpenditure := Budget{0.0, 0.0}

	transactions := []Transaction{}

	trialBudget1 := TrialBudget{trialBudget, trialExpenditure, transactions}
	// Fetching the site budget JSON for variable cost
	sites := protocolMap["sites"].([]interface{})
	siteBudgets := []SiteBudget{}

	/* Bifurcating the site budget fixed and variable costs along with various factors
	summing up to the fixed and variable costs.*/
	for _, element := range sites {
		siteMap := element.(map[string]interface{})
		siteDetails := siteMap["site_details"].(map[string]interface{})
		siteKey := siteDetails["site_id"].(float64)
		budget := siteMap["site_budget"].(map[string]interface{})
		fixed := budget["fixed_cost"].(map[string]interface{})
		siteFixedMilestoneCost := fixed["milestone_cost"].(map[string]interface{})
		siteFixedMilestone := MilestoneCosts{siteFixedMilestoneCost["cost"].(float64),
			siteFixedMilestoneCost["comment"].(string)}

		// Initializing the site fixed cost
		siteFixedCost := SiteFixed{fixed["courier_costs"].(float64),
			fixed["irb_fees"].(float64), siteFixedMilestone,
			fixed["total_cost"].(float64), fixed["regulatory_fees"].(float64),
			fixed["manpower_costs"].(float64)}

		variable := budget["variable_cost"].(map[string]interface{})
		siteVariableMilestoneCost := variable["milestone_cost"].(map[string]interface{})
		siteVariableMilestone := MilestoneCosts{siteVariableMilestoneCost["cost"].(float64),
			siteVariableMilestoneCost["comment"].(string)}

		// Initializing the site variable cost
		siteVariableCost := SiteVariable{variable["courier_costs"].(float64),
			siteVariableMilestone, variable["total_cost"].(float64),
			variable["ae_sae_up_cost"].(float64), variable["patient_visits"].(float64),
			variable["diagnostics"].(float64), variable["travel_cost_patient"].(float64),
			variable["imaging"].(float64)}

		// Initializing the site budget including the fixed and variable costs
		siteBudget := SiteBudgetDetails{fixed["total_cost"].(float64),
			variable["total_cost"].(float64), siteFixedCost, siteVariableCost}

		// Initializing the site expenditure including the fixed and variable costs
		siteExpenditure := Budget{0.0, 0.0}

		temp := SiteBudget{siteKey, siteBudget, siteExpenditure, transactions}
		siteBudgets = append(siteBudgets, temp)
	}
	budget := TotalBudget{protocolID, trialBudget1, siteBudgets}
	budgetJSON, _ := json.Marshal(budget)

	// Encrypts and stores the budget data in a new block
	err := encryptAndPutState(stub, t.bccspInst, transientMap, protocolID+"_bud", []byte(budgetJSON))
	if err != nil {
		return false, errors.New("Failed to create budget")
	}

	return true, nil
}

/**
 * @Description
 *	- Accepts the trial id, queries the blockchain and fetches the record respective to that trial.
 *
 * @params
 *	- []string: args
 *		- args is a string array containing the id of the trial to be fetched.
 *			- args[0]: id of the trial for which record is to be fetched.
 *
 * @returns
 *	- payload: []byte
 *		- returns byte array containing the trial details corresponding to that trial id.
 *
 * @returns
 *	- payload: string
 *		- returns error if the trial record was not fetched.
 *			- Incorrect number of arguements passed: "Incorrect number of arguments.
 *				Expecting protocol_id"
 *			- Trial query failed: "Error {error} (Error returned from the blockchain)"
 *
 * @changelog
 */

func (t *SimpleChaincode) queryTrial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Trial ===")

	//Entities
	var protocolID string
	var err error

	// Check whether trial id is present
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting protocol_id")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	protocolID = args[0]

	// Decrypt and get the trial data corresponding to that trial id.
	protocolValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, protocolID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error %s", err))
	}

	return shim.Success(protocolValBytes)
}

/**
 * @Description
 *	- Accepts the site selection questionnaire id and the site selection questionnaire details
 *		and creates a new block storing that information against that particular site selection
 *		questionnaire id.
 *
 * @params
 *	- []string: args
 *		- args is a string array containing the new site selection questionnaire id and the
 *			corresponding site selection questionnaire details.
 *			- args[0]: id of the new site selection questionnaire.
 *			- args[1]: site selection questionnaire details to be stored.
 *				protocol_id:
 *                type: string
 *                required: true
 *              device_uuid:
 *                type: string
 *                required: true
 *              location:
 *                type: object
 *                required: true
 *                properties:
 *                  lat:
 *                    type: string
 *                    required: true
 *                  long:
 *                    type: string
 *                    required: true
 *              site:
 *                type: object
 *                required: true
 *                properties:
 *                  site_details:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      site_name:
 *                        type: string
 *                        required: true
 *                      site_id:
 *                        type: number
 *                        required: true
 *                  department_name:
 *                    type: string
 *                    required: true
 *                  practice_type:
 *                    type: string
 *                    required: true
 *                  active_patients_count:
 *                    type: number
 *                    required: true
 *                  capable_physicians_count:
 *                    type: number
 *                    required: true
 *                  active_studies_count:
 *                    type: number
 *                    required: true
 *                  physicians_as_pi_count:
 *                    type: number
 *                    required: true
 *              principal_investigator:
 *                type: object
 *                required: true
 *                properties:
 *                  investigator_details:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      user_id:
 *                        type: string
 *                        format: uuid
 *                        required: true
 *                      name:
 *                        type: string
 *                        required: true
 *                      credentials:
 *                        type: string
 *                        required: true
 *                      address:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          line1:
 *                            type: string
 *                            required: true
 *                          line2:
 *                            type: string
 *                            required: true
 *                      website:
 *                        type: string
 *                        required: true
 *                      phone:
 *                        type: string
 *                        required: true
 *                      mobile:
 *                        type: string
 *                        required: true
 *                      fax:
 *                        type: string
 *                        required: true
 *                      time_and_resource_available:
 *                        type: boolean
 *                        required: true
 *                      consent_to_conduct_study:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: boolean
 *                            required: true
 *                          comments:
 *                            type: string
 *                      primary_practice_areas:
 *                        type: number
 *                        required: true
 *                      time_devotion_for_research:
 *                        type: number
 *                        required: true
 *                      training_and_certifications:
 *                        type: string
 *                        required: true
 *                      site_network_affiliation:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: boolean
 *                            required: true
 *                          comments:
 *                            type: string
 *                      research_conducted_in_three_years:
 *                        type: number
 *                        required: true
 *                      industry_sponsored_studies:
 *                        type: number
 *                        required: true
 *                      pi_in_studies:
 *                        type: number
 *                        required: true
 *                      research_studies_conducting_count:
 *                        type: number
 *                        required: true
 *                      research_studies_in_three_years:
 *                        type: number
 *                        required: true
 *                      research_studies_in_three_years_on_relevant_conditions:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: number
 *                            required: true
 *                          comments:
 *                            type: string
 *                      key_opinion_leader:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: number
 *                            required: true
 *                          comments:
 *                            type: string
 *                      research_experience:
 *                        type: number
 *                        required: true
 *                      studies_conducted_per_year:
 *                        type: number
 *                        required: true
 *                      investigator_initiated_studies:
 *                        type: number
 *                        required: true
 *                      last_study_date:
 *                        type: string
 *                        format: date
 *                        required: true
 *                      compound_class:
 *                        type: string
 *                        required: true
 *                      indication:
 *                        type: string
 *                        required: true
 *                      phase:
 *                        type: string
 *                        required: true
 *                      enrolled_subjects_count:
 *                        type: number
 *                        required: true
 *                      similar_study_experience:
 *                        type: boolean
 *                        required: true
 *                      board_certified:
 *                        type: boolean
 *                        required: true
 *                      dea_license_available:
 *                        type: boolean
 *                        required: true
 *                      study_conducted_with_schedule_one_study_drug:
 *                        type: boolean
 *                        required: true
 *              study_coordinator:
 *                type: object
 *                properties:
 *                  active_enrolling_studies_managed_by_crc:
 *                    type: number
 *                    required: true
 *                  participating_coordinators_count:
 *                    type: number
 *                    required: true
 *                  primary_study_coordinator_details:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      name:
 *                        type: string
 *                        required: true
 *                      crc_at_site_time_period:
 *                        type: string
 *                        format: date
 *                        required: true
 *                      research_experience_time_period:
 *                        type: string
 *                        required: true
 *                      credentials:
 *                        type: string
 *                        required: true
 *                      training_and_certifications:
 *                        type: string
 *                        required: true
 *                      work_time_spent_on_research:
 *                        type: number
 *                        required: true
 *                      relevant_studies_managed:
 *                        type: number
 *                        required: true
 *                      time_available_for_study:
 *                        type: boolean
 *                        required: true
 *                      want_to_manage_study:
 *                        type: boolean
 *                        required: true
 *                      contact_info:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          phone:
 *                            type: string
 *                            required: true
 *                          fax:
 *                            type: string
 *                            required: true
 *                          email:
 *                            type: string
 *                            required: true
 *                      edc_experience_studies:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: boolean
 *                            required: true
 *                          comments:
 *                            type: string
 *                      high_speed_internet_access_for_edc:
 *                        type: boolean
 *                        required: true
 *                      experience_with_electronic_subject_diaries:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          value:
 *                            type: boolean
 *                            required: true
 *                          comments:
 *                            type: string
 *                  other_key_members:
 *                    type: array
 *                    required: true
 *                    items:
 *                      type: object
 *                      required: true
 *                      properties:
 *                        name:
 *                          type: string
 *                          required: true
 *                        role:
 *                          type: string
 *                          required: true
 *              capabilities_and_resources:
 *                type: object
 *                required: true
 *                properties:
 *                  adequate_experience_with_assessment:
 *                    type: boolean
 *                    required: true
 *                  person_performing_assessment:
 *                    type: string
 *                    required: true
 *                  access_to_lab_with_adequate_capacity:
 *                    type: string
 *                    required: true
 *                  actively_use_sops:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  sub_investigator_enrolling_subjects_count:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: number
 *                        required: true
 *                      comments:
 *                        type: string
 *                  usage_of_payment_and_budgeting_systems:
 *                    type: boolean
 *                    required: true
 *                  use_of_central_laboratory:
 *                    type: boolean
 *                    required: true
 *                  availability_of_local_laboratory:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  site_access_to_public_transportation:
 *                    type: string
 *                    required: true
 *                  parking_availability:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  waiting_area_available_for_family_members:
 *                    type: string
 *                    required: true
 *                  availability_of_food_and_bevarage:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: string
 *                        required: true
 *                      comments:
 *                        type: string
 *                  availibility_of_pharmacist_to_prepare_drugs:
 *                    type: boolean
 *                    required: true
 *                  sub_investigator_for_blinded_assessments_availibility:
 *                    type: boolean
 *                    required: true
 *                  brand_and_model_of_equipment:
 *                    type: string
 *                    required: true
 *                  administer_study_drug:
 *                    type: string
 *                    required: true
 *                  dedicated_fax_line:
 *                    type: boolean
 *                    required: true
 *                  internet_access_to_monitors_without_login:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  participanting_satellite_sites_count:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: number
 *                        required: true
 *                      comments:
 *                        type: string
 *              site_initiation:
 *                type: object
 *                required: true
 *                properties:
 *                  time_taken_by_site_to_start_study:
 *                    type: string
 *                    required: true
 *                  site_pursue_approval_at_same_time_as_contract_completion:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: number
 *                        required: true
 *                      comments:
 *                        type: string
 *                  use_irb_for_current_study:
 *                    type: boolean
 *                    required: true
 *                  name_of_another_irb_to_use:
 *                    type: string
 *                    required: true
 *                  how_often_does_it_meet:
 *                    type: string
 *                    required: true
 *                  time_taken_from_submission_to_approval_letter:
 *                    type: string
 *                    required: true
 *                  additional_approval_or_review_committee:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  use_of_standard_trial_agreement_template:
 *                    type: boolean
 *                    required: true
 *                  contracts_needed_other_than_trial_agreement:
 *                    type: string
 *                    required: true
 *                  attend_site_qualification_visit:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  investigator_attend_investigator_meetings:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *              subject_enrollment:
 *                type: object
 *                required: true
 *                properties:
 *                  enrolled_subject_count_by_investigator_in_12_months:
 *                    type: number
 *                    required: true
 *                  enrolled_subject_count_by_investigator_in_3_years_for_relevant_study:
 *                    type: number
 *                    required: true
 *                  number_of_relevant_studies_enrolling_subjects:
 *                    type: string
 *                    required: true
 *                  active_patients_investigator_have:
 *                    type: number
 *                    required: true
 *                  subject_count_diagnosed_till_date:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: number
 *                        required: true
 *                      comments:
 *                        type: string
 *                  subject_count_meeting_eligibility_criteria:
 *                    type: number
 *                    required: true
 *                  expected_enrollment_subject_count_based_on_study_summary:
 *                    type: string
 *                    required: true
 *                  expected_additional_enrollment_subject_count_from_site:
 *                    type: number
 *                    required: true
 *                  expected_additional_enrollment_subject_count_outside_site:
 *                    type: number
 *                    required: true
 *                  screening_patients_count_for_each_enrollment:
 *                    type: number
 *                    required: true
 *                  study_appeal_to_patients:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  site_use_emr:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      patients_count_in_emr:
 *                        type: number
 *                        required: true
 *                      accessibility_of_records_by_monitor:
 *                        type: string
 *                        required: true
 *                  potential_subjects_database:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                      active_patients_count:
 *                        type: number
 *                        required: true
 *                  method_to_find_subject_outside_site:
 *                    type: string
 *                    required: true
 *                  method_to_contact_potential_subjects:
 *                    type: string
 *                    required: true
 *                  informed_consent_language_other_than_english:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      comments:
 *                        type: string
 *                  estimated_monthly_subject_count:
 *                    type: string
 *                    required: true
 *                  patient_interested_in_participation:
 *                    type: number
 *                    required: true
 *                  study_offer_pattients_acceptable_risk_benifit_ratio:
 *                    type: boolean
 *                    required: true
 *                  type_of_patients_to_enroll_or_not_enroll:
 *                    type: string
 *                    required: true
 *                  patients_participant_in_clinical_trials:
 *                    type: number
 *                    required: true
 *                  method_to_identify_potential_subjects:
 *                    type: string
 *                    required: true
 *                  method_to_recruit_subjects_for_studies:
 *                    type: string
 *                    required: true
 *                  method_to_recruit_subject_for_current_study:
 *                    type: string
 *                    required: true
 *                  contingency_plan_if_initial_recruiting_plan_is_inadequate:
 *                    type: string
 *                    required: true
 *                  backup_plan_if_initial_recruiting_plan_is_inadequate:
 *                    type: string
 *                    required: true
 *                  willing_to_submit_blinded_list:
 *                    type: boolean
 *                    required: true
 *                  willing_to_work_with_central_subject_recruiting_program:
 *                    type: boolean
 *                    required: true
 *                  access_to_potential_subjects_in_facility:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      travelling_problem:
 *                        type: boolean
 *                        required: true
 *                  subject_recruitment_site_department:
 *                    type: string
 *                    required: true
 *                  enroll_potential_subjects_24/7:
 *                    type: boolean
 *                    required: true
 *                  serving_geographical_area_with_population:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      area:
 *                        type: string
 *                        required: true
 *                      population:
 *                        type: number
 *                        required: true
 *                  details_of_potential_subjects:
 *                    type: string
 *                    required: true
 *                  patients_received_chemotherapy_previously:
 *                    type: number
 *                    required: true
 *                  standard_of_care:
 *                    type: string
 *                    required: true
 *                  patients_normally_receiving_chemotherapy:
 *                    type: number
 *                    required: true
 *                  willing_and_able_to_use_chemotherapy:
 *                    type: boolean
 *                    required: true
 *                  currently_using_chemotherapy_for_treatment:
 *                    type: boolean
 *                    required: true
 *                  willing_and_able_to_conduct_placebo_controlled_study:
 *                    type: boolean
 *                    required: true
 *              governance_and_compliance:
 *                type: object
 *                required: true
 *                properties:
 *                  site_inspected_by_regulatory_agency_in_past_five_years:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      outcome:
 *                        type: string
 *                        required: true
 *                      document_url:
 *                        type: number
 *                        required: true
 *                  challenges_and_risk_for_study:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: string
 *                        required: true
 *                      method_to_address:
 *                        type: string
 *                        required: true
 *                      method_to_address_by_sponsor:
 *                        type: string
 *                        required: true
 *                  additional_comments:
 *                    type: string
 *                    required: true
 *                  uncertain_answers:
 *                    type: string
 *                    required: true
 *                  person_completing_questionnaire:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      user_id:
 *                        type: string
 *                        format: uuid
 *                        required: true
 *                      name:
 *                        type: string
 *                        required: true
 *                      role:
 *                        type: object
 *                        required: true
 *                        properties:
 *                          role_id:
 *                            type: number
 *                            required: true
 *                          title:
 *                            type: string
 *                            required: true
 *                      contact_info:
 *                        type: string
 *                        required: true
 *              others:
 *                type: object
 *                required: true
 *                properties:
 *                  site_audited_by_sponsor_in_past_five_years:
 *                    type: object
 *                    required: true
 *                    properties:
 *                      value:
 *                        type: boolean
 *                        required: true
 *                      time_and_result:
 *                        type: string
 *                        required: true
 *                  participation_in_sampling_part_of_study:
 *                    type: boolean
 *                    required: true
 *                  experience_with_sampling:
 *                    type: boolean
 *                    required: true
 *                  primary_language_at_site:
 *                    type: string
 *                    required: true
 *                  investigator_and_crf_speak_english:
 *                    type: boolean
 *                    required: true
 *                  24/7_contact_available_for_subjects:
 *                    type: boolean
 *                    required: true
 *
 * @returns
 *	- payload: nil
 *		- returns nil if the site selection questionnaire record is successfully created.
 *
 * @returns
 *	- payload: string
 *		- returns error if the site selection questionnaire record creation was unsuccessful.
 *			- Incorrect number of arguements passed: "Incorrect number of arguments.
 *				Expecting Key-Value pair"
 *			- Site selection questionnaire record creation failed: "Failed to create
 *				site selection questionnaire"
 *
 * @changelog
 */

func (t *SimpleChaincode) createSiteSelectionQue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Site Sel Que ===")

	// Check the number of arguements required for site selection questionnaire creation
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting Key-Value pair")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	// Entities
	queID := args[0]
	queDetails := args[1]

	// Encrypts and stores the site selection questionnaire data in a new block
	err = encryptAndPutState(stub, t.bccspInst, transientMap, queID, []byte(queDetails))
	if err != nil {
		return shim.Error("Failed to create site selection questionnaire")
	}

	return shim.Success(nil)
}

/**
* @Description
*	- Accepts the site selection questionnaire id, queries the blockchain and
*		fetches the record respective to that trial and that particular site.
*
* @params
*	- []string: args
*		- args is a string array containing the id of the site selection questionnaire
*			for which the record is to be fetched.
*			- args[0]: id of the site selection questionnaire for which record is to be fetched.
*
* @returns
*	- payload: []byte
*		- returns byte array containing the site selection questionnaire details corresponding
*			to that site selection questionnaire id.
*
* @returns
*	- payload: string
*		- returns error if the site selection questionnaire record was not fetched.
*			- Incorrect number of arguements passed: "Incorrect number of arguments.
*				Expecting site selection questionnaire id"
*			- Site selection questionnaire query failed: "Failed to get state for
*				{site selection questionnaire id}"
*			- No record for site selection questionnaire: "Nil value for {site selection
*				questionnaire id}"
*
* @changelog
 */

func (t *SimpleChaincode) querySiteSelectionQue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Site Sel Que ===")

	//Entities
	var queID string
	var err error

	// Check whether site selection questionnaire id is present
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting site selection questionnaire id")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	queID = args[0]

	/* Decrypt and get the site selection questionnaire data corresponding to that site selection
	questionnaire id.*/
	queValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, queID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + queID + "\"}"
		return shim.Error(jsonResp)
	}

	// Check if there is no record against that site selection questionnaire id
	if queValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + queID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(queValBytes)
}

/**
* @Description
*	- Accepts the pre site visit id and the pre site visit details and creates a
*		new block storing that information against that particular pre site visit id.
*
* @params
*	- []string: args
*		- args is a string array containing the new pre site visit id and the
*			corresponding pre site visit details.
*			- args[0]: id of the new pre site visit record.
*			- args[1]: pre site visit details to be stored.
*				investigator_name:
*       		  type: string
*       		site_address:
*       		  type: string
*       		contact_information:
*       		  type: string
*       		study_type:
*       		  type: string
*       		protocol_id:
*       		  type: string
*       		  required: true
*       		site_id:
*       		  type: integer
*       		  required: true
*       		site_visit_date:
*       		  type: string
*       		  format: date-time
*       		site_personnel:
*       		  type: array
*       		  items:
*       		    type: object
*       		    properties:
*       		      name:
*       		        type: string
*       		      role:
*       		        type: string
*       		SM_representative:
*       		  type: array
*       		  items:
*       		    type: object
*       		    properties:
*       		      name:
*       		        type: string
*       		      role:
*       		        type: string
*       		investigator_experience:
*       		  type: integer
*       		past_experience:
*       		  type: boolean
*       		completed_trials:
*       		  type: integer
*       		ongoing_trials:
*       		  type: integer
*       		ongoing_competitive_trials:
*       		  type: integer
*       		favourable_points:
*       		  type: string
*       		inadequacy:
*       		  type: string
*       		visit_to_be_conducted:
*       		  type: boolean
*       		visit_conducted:
*       		  type: boolean
*       		site_recommended:
*       		  type: boolean
*       		site_suitable_for_others:
*       		  type: boolean
*       		SM_representative_signature:
*       		  type: object
*       		  properties:
*       		    signature:
*       		      type: string
*       		    name:
*       		      type: string
*       		    date:
*       		      type: string
*       		review_and_approval:
*       		  type: object
*       		  properties:
*       		    comment:
*       		      type: string
*       		    reviewer_signature:
*       		      type: string
*       		    reviewer_name:
*       		      type: string
*       		    review_date:
*       		      type: string
*       		      format: date-time
*       		    QAU_signature:
*       		      type: string
*       		    QAU_name:
*       		      type: string
*       		    QAU_approval_date:
*       		      type: string
*       		      format: date-time
*
* @returns
*	- payload: nil
*		- returns nil if the pre site visit record is successfully created.
*
* @returns
*	- payload: string
*		- returns error if the pre site visit record creation was unsuccessful.
*			- Incorrect number of arguements passed: "Incorrect number of arguments."
*			- Pre site visit record creation failed: "Failed to create
*				pre site visit"
*
* @changelog
 */

func (t *SimpleChaincode) createPreSiteVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Pre Site Visit ===")

	// Check the number of arguements required for pre site visit record creation
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	// Entities
	preSiteVisitID := args[0]
	preSiteVisitDetails := args[1]

	// Encrypts and stores the pre site visit data in a new block
	err = encryptAndPutState(stub, t.bccspInst, transientMap, preSiteVisitID, []byte(preSiteVisitDetails))
	if err != nil {
		return shim.Error("Failed to create Pre Site Visit")
	}

	return shim.Success(nil)
}

/**
* @Description
*	- Accepts the pre site visit id, queries the blockchain and
*		fetches the record respective to that trial and that particular site.
*
* @params
*	- []string: args
*		- args is a string array containing the id of the pre site visit
*			for which the record is to be fetched.
*			- args[0]: id of the pre site visit for which record is to be fetched.
*
* @returns
*	- payload: []byte
*		- returns byte array containing the pre site visit details corresponding
*			to that pre site visit id.
*
* @returns
*	- payload: string
*		- returns error if the pre site visit record was not fetched.
*			- Incorrect number of arguements passed: "Incorrect number of arguments.
*				Expecting pre site visit id"
*			- Pre site visit query failed: "Failed to get state for
*				{pre site visit id}"
*			- No record for pre site visit: "Nil value for {pre site visit id}"
*
* @changelog
 */

func (t *SimpleChaincode) queryPreSiteVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Pre Site Visit ===")

	// Entities
	var preSiteVisitID string
	var err error

	// Check whether pre site visit id is present
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting pre_site_visit_id")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	preSiteVisitID = args[0]

	// Decrypt and get the pre site visit data corresponding to that pre site visit record id.
	preSiteVisitValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, preSiteVisitID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + preSiteVisitID + "\"}"
		return shim.Error(jsonResp)
	}

	// Check if there is no record against that pre site visit id
	if preSiteVisitValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + preSiteVisitID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(preSiteVisitValBytes)
}

/**
* @Description
*	- Accepts the site selection visit checklist id and the site selection visit checklist
*		details and creates a new block storing that information against that particular
*		site selection visit checklist id.
*
* @params
*	- []string: args
*		- args is a string array containing the new site selection visit checklist id and the
*			corresponding site selection visit checklist details.
*			- args[0]: id of the new site selection visit checklist record.
*			- args[1]: site selection visit checklist details to be stored.
*				sponsor_id:
*       		  type: string
*       		site_id:
*       		  type: integer
*       		  required: true
*       		protocol_id:
*       		  type: string
*       		  required: true
*       		prep_for_visit:
*       		  type: object
*       		  required: true
*       		  properties:
*       		    sponsor_representatives:
*       		      type: array
*       		      items:
*       		        type: object
*       		        properties:
*       		          representative_id:
*       		            type: string
*       		          title:
*       		            type: string
*       		          phone:
*       		            type: string
*       		          mobile:
*       		            type: string
*       		          email:
*       		            type: string
*       		          fax:
*       		            type: string
*       		    visit_date:
*       		      type: string
*       		      format: date-time
*       		    site_attendees_available:
*       		      type: boolean
*       		    departments_aware_of_visit:
*       		      type: boolean
*       		    docs_circulated:
*       		      type: object
*       		      properties:
*       		        protocol:
*       		          type: boolean
*       		        investigation_brochure:
*       		          type: boolean
*       		        therapeutic_agent:
*       		          type: boolean
*       		    meeting_room:
*       		      type: object
*       		      properties:
*       		        AV_requirements:
*       		          type: boolean
*       		        refreshments:
*       		          type: boolean
*       		    dept_meeting_dates:
*       		      type: object
*       		      properties:
*       		        IRB_meeting_date:
*       		          type: string
*       		          format: date-time
*       		        IRB_deadline:
*       		          type: string
*       		          format: date-time
*       		        institution_meeting_date:
*       		          type: string
*       		          format: date-time
*       		        institution_deadline:
*       		          type: string
*       		          format: date-time
*       		    documents_compiled:
*       		      type: array
*       		      items:
*       		        type: object
*       		        properties:
*       		          pi_cv:
*       		            type: boolean
*       		          coordinator_cv:
*       		            type: boolean
*       		          sop_index:
*       		            type: boolean
*       		          souce_doc_example:
*       		            type: boolean
*       		          recruitment_potential:
*       		            type: boolean
*       		          past_trial_evidence:
*       		            type: boolean
*       		    signature:
*       		      type: object
*       		      properties:
*       		        signature:
*       		          type: string
*       		        user_id:
*       		          type: string
*       		        date:
*       		          type: string
*       		          format: date-time
*       		site_selection_visit_summary:
*       		  type: object
*       		  properties:
*       		    sponsor_representatives:
*       		      type: array
*       		      items:
*       		        type: object
*       		        properties:
*       		          representative_id:
*       		            type: string
*       		          title:
*       		            type: string
*       		          phone:
*       		            type: string
*       		          mobile:
*       		            type: string
*       		          email:
*       		            type: string
*       		          fax:
*       		            type: string
*       		    attendees:
*       		      type: array
*       		      items:
*       		        type: object
*       		        properties:
*       		          attendee_id:
*       		            type: string
*       		          title:
*       		            type: string
*       		          phone:
*       		            type: string
*       		          mobile:
*       		            type: string
*       		          email:
*       		            type: string
*       		          fax:
*       		            type: string
*       		    summary_of_meeting:
*       		      type: object
*       		      properties:
*       		        key_points:
*       		          type: string
*       		        agenda:
*       		          type: boolean
*       		        meeting_minutes:
*       		          type: boolean
*       		    items_reviewed:
*       		      type: object
*       		      properties:
*       		        protocol:
*       		          type: boolean
*       		        investigational_product:
*       		          type: boolean
*       		        study_monitoring_plans:
*       		          type: boolean
*       		        data_management_plans:
*       		          type: boolean
*       		        publication_policy:
*       		          type: boolean
*       		        anticipated_dates:
*       		          type: object
*       		          properties:
*       		            regulatory_approval:
*       		              type: string
*       		              format: date-time
*       		            overall_study_start:
*       		              type: string
*       		              format: date-time
*       		            site_study_start:
*       		              type: string
*       		              format: date-time
*       		            investigator_meeting:
*       		              type: string
*       		              format: date-time
*       		            accural_end:
*       		              type: string
*       		              format: date-time
*       		            site_selection_notification:
*       		              type: string
*       		              format: date-time
*       		    tour_of_facilities:
*       		      type: object
*       		      properties:
*       		        pharmacy:
*       		          type: boolean
*       		        laboratory:
*       		          type: boolean
*       		        pathology_facilities:
*       		          type: boolean
*       		        clinics:
*       		          type: boolean
*       		        monitor_work_area:
*       		          type: boolean
*       		        investigational_drug_storage_area:
*       		          type: boolean
*       		        radiology_and_imaging_facilities:
*       		          type: boolean
*       		        examination_rooms:
*       		          type: boolean
*       		        surgical_rooms:
*       		          type: boolean
*       		        treatment_area:
*       		          type: object
*       		          properties:
*       		            treatment_area:
*       		              type: boolean
*       		            comment:
*       		              type: string
*       		        other:
*       		          type: object
*       		          properties:
*       		            other:
*       		              type: boolean
*       		            comment:
*       		              type: string
*       		    action_items:
*       		      type: array
*       		      items:
*       		        type: object
*       		        properties:
*       		          action_item:
*       		            type: string
*       		          responsible_person_id:
*       		            type: string
*       		          due_date:
*       		            type: string
*       		            format: date-time
*       		    follow_up_required:
*       		      type: object
*       		      properties:
*       		        follow_up_1_comment:
*       		          type: string
*       		        follow_up_2_comment:
*       		          type: string
*       		    outcome:
*       		      type: object
*       		      properties:
*       		        site_selected:
*       		          type: boolean
*       		        site_agrees_to_conduct:
*       		          type: boolean
*       		        site_declined:
*       		          type: object
*       		          properties:
*       		            site_declined:
*       		              type: boolean
*       		            comment:
*       		              type: string
*       		        sponsor_declined:
*       		          type: object
*       		          properties:
*       		            sponsor_declined:
*       		              type: boolean
*       		            comment:
*       		              type: string
*       		    signature_of_person_completing_form:
*       		      type: object
*       		      properties:
*       		        signature:
*       		          type: string
*       		        user_id:
*       		          type: string
*       		        date:
*       		          type: string
*       		          format: date-time
*
* @returns
*	- payload: nil
*		- returns nil if the site selection visit checklist record is successfully created.
*
* @returns
*	- payload: string
*		- returns error if the site selection visit checklist record creation was unsuccessful.
*			- Incorrect number of arguements passed: "Incorrect number of arguments."
*			- Site selection visit checklist record creation failed: "Failed to create
*				site selection visit checklist"
*
* @changelog
 */

func (t *SimpleChaincode) createSiteSelectionVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Site Selection Visit ===")

	// Check the number of arguements required for site selection visit checklist record creation
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	// Entities
	siteSelectionVisitID := args[0]
	siteSelectionVisitDetails := args[1]

	// Encrypts and stores the site selection visit checklist data in a new block
	err = encryptAndPutState(stub, t.bccspInst, transientMap, siteSelectionVisitID, []byte(siteSelectionVisitDetails))
	if err != nil {
		return shim.Error("Failed to create Site Visit Selection Checklist")
	}

	return shim.Success(nil)
}

/**
* @Description
*	- Accepts the site selection visit checklist id, queries the blockchain and
*		fetches the record respective to that trial and that particular site.
*
* @params
*	- []string: args
*		- args is a string array containing the id of the site selection visit checklist
*			for which the record is to be fetched.
*			- args[0]: id of the site selection visit checklist for which record is to be fetched.
*
* @returns
*	- payload: []byte
*		- returns byte array containing the site selection visit checklist details corresponding
*			to that site selection visit checklist id.
*
* @returns
*	- payload: string
*		- returns error if the site selection visit checklist record was not fetched.
*			- Incorrect number of arguements passed: "Incorrect number of arguments.
*				Expecting site selection visit id"
*			- Site selection visit checklist query failed: "Failed to get state for
*				{site selection visit checklist id}"
*			- No record for site selection visit checklist: "Nil value for {site selection visit
*	 			checklist id}"
*
* @changelog
 */

func (t *SimpleChaincode) querySiteSelectionVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Site Selection Visit ===")

	// Entities
	var siteSelectionVisitID string
	var err error

	// Check whether site selection visit checklist id is present
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting site_selection_visit_id")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	siteSelectionVisitID = args[0]

	/* Decrypt and get the site selection visit checklist data corresponding to that site
	selection visit checklist record id.*/
	siteSelectionVisitValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, siteSelectionVisitID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + siteSelectionVisitID + "\"}"
		return shim.Error(jsonResp)
	}

	// Check if there is no record against that site selection visit checklist id
	if siteSelectionVisitValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + siteSelectionVisitID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(siteSelectionVisitValBytes)
}

/**
 * @Description
 *	- Fetches budget according to that particular trial id.
 *
 * @params
 *	- []string: args
 *		- args is a string array containing the id of the trial for which the budget
 *			record is to be fetched.
 *			- args[0]: id of the trial for which the budget is to be fetched.
 *
 * @returns
 *	- payload: []byte
 *		- returns byte array containing the budget details for that trial.
 *
 * @returns
 *	- payload: string
 *		- returns error if the fetching budget record was not fetched.
 *			- Incorrect number of arguements passed: "Incorrect number of arguments.
 *				Expecting protocol id"
 *			- Budget query failed: "Failed to get state for {budget id}".
 *
 * @changelog
 */

func (t *SimpleChaincode) queryBudget(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//Entities
	var protocolID string
	var err error

	// Check whether trial id is present
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting protocol_id")
	}

	// Check if the provided encryption key and IV are correct
	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	protocolID = args[0]

	// Decrypt and get the budget data corresponding to that trial id
	budgetValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, protocolID+"_bud")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + protocolID + "\"}"
		return shim.Error(jsonResp)
	}

	if budgetValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + protocolID + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(budgetValBytes)
}

//SiteInitiate -
func (t *SimpleChaincode) createSiteInitiate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments.")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	siteInitiationID := args[2]
	siteInitiationDetails := args[3]

	var siteInitationJSON interface{}
	json.Unmarshal([]byte(siteInitiationDetails), &siteInitationJSON)

	siteInitationMap := siteInitationJSON.(map[string]interface{})
	documentsOrActivity := siteInitationMap["documents_or_activity"].(map[string]interface{})

	flag := true
	for _, element := range documentsOrActivity {
		val := element.(map[string]interface{})
		if val["value"] == "No" {
			flag = false
		}
	}

	if !flag {
		err := encryptAndPutState(stub, t.bccspInst, transientMap, siteInitiationID, []byte(siteInitiationDetails))
		if err != nil {
			return shim.Error("Failed to initiate site")
		}

		jsonResp := "{\"Error\":\"Site cannot initiated as one of the condition is not matched\"}"
		return shim.Error(jsonResp)
	} else {
		err = encryptAndPutState(stub, t.bccspInst, transientMap, siteInitiationID, []byte(siteInitiationDetails))
		if err != nil {
			return shim.Error("Failed to initiate site")
		}
	}

	return shim.Success(nil)
}

//QuerySiteInitiate -
func (t *SimpleChaincode) querySiteInitiate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Site Initiate ===")
	var siteInitiationID string //Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting site_initiation_id")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	siteInitiationID = args[0]

	siteInitiateBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, siteInitiationID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + siteInitiationID + "\"}"
		return shim.Error(jsonResp)
	}

	if siteInitiateBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + siteInitiationID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(siteInitiateBytes)
}

//UpdateBudget -
func (t *SimpleChaincode) updateBudget(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== UpdateBudget ===")
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	chargeType := args[0]
	protocolID := args[1]
	siteID := args[2]
	charge, _ := strconv.ParseFloat(args[3], 64)
	transaction := args[4]

	var transStruct Transaction
	json.Unmarshal([]byte(transaction), &transStruct)

	// Get the state from the ledger
	budgetValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, protocolID+"_bud")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + protocolID + "\"}"
		return shim.Error(jsonResp)
	}

	var totalBudget TotalBudget
	json.Unmarshal(budgetValBytes, &totalBudget)

	siteBudgets := []SiteBudget{}
	for _, element := range totalBudget.SiteBudget {
		if strconv.FormatFloat(element.Site_ID, 'f', 0, 64) == siteID {
			if chargeType == "fixed" {
				element.Expenditure.Fixed_Cost = element.Expenditure.Fixed_Cost + charge
			} else if chargeType == "variable" {
				element.Expenditure.Variable_Cost = element.Expenditure.Variable_Cost + charge
			}
			element.Transactions = append(element.Transactions, transStruct)
		}

		siteBudgets = append(siteBudgets, element)
	}

	budget := TotalBudget{totalBudget.Protocol_ID, totalBudget.TrialBudget, siteBudgets}

	budgetJSON, _ := json.Marshal(budget)

	err = encryptAndPutState(stub, t.bccspInst, transientMap, protocolID+"_bud", []byte(budgetJSON))

	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to update budget, err %s", err))
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) createSubject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Subject ===")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	subjectID := args[0]
	subjectDetails := args[1]

	err = encryptAndPutState(stub, t.bccspInst, transientMap, subjectID, []byte(subjectDetails))
	if err != nil {
		return shim.Error("Failed to create subject")
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) querySubject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Subject ===")
	var subjectID string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting subject_id")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	subjectID = args[0]

	subjectValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, subjectID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + subjectID + "\"}"
		return shim.Error(jsonResp)
	}

	if subjectValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + subjectID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(subjectValBytes)
}

func (t *SimpleChaincode) createCrf(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Crf ===")
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments.")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	crfID := args[0]
	crfDetails := args[1]

	err = encryptAndPutState(stub, t.bccspInst, transientMap, crfID, []byte(crfDetails))
	if err != nil {
		return shim.Error("Failed to create crf")
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) queryCrf(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Crf ===")
	var crfID string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting crf_id")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	crfID = args[0]

	crfValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, crfID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + crfID + "\"}"
		return shim.Error(jsonResp)
	}

	if crfValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + crfID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(crfValBytes)
}

func (t *SimpleChaincode) getHistoryForCrf(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	crfID := args[0]

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	resultsIterator, err := stub.GetHistoryForKey(crfID)

	if err != nil {
		return shim.Error(err.Error())
	}

	// buffer is a JSON array containing historic values for the crf
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON crf)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			decryptedValue, err := GetHistoryForKeyAndDecrypt(stub, t.bccspInst, transientMap, string(response.Value))
			if err != nil {
				return shim.Error(err.Error())
			}
			buffer.WriteString(string(decryptedValue))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

// store protocol deviation
func (t *SimpleChaincode) createProtocolDeviation(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Create Protocol deviation ===")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	protocolDeviationID := args[0]
	protocolDeviationDetails := args[1]

	err = encryptAndPutState(stub, t.bccspInst, transientMap, protocolDeviationID, []byte(protocolDeviationDetails))
	if err != nil {
		return shim.Error("Failed to create protocol deviation")
	}

	return shim.Success(nil)
}

// query protocol deviation
func (t *SimpleChaincode) queryProtocolDeviation(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("=== Query Protocol deviation ===")
	var protocolDeviationID string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting subject_id")
	}

	transientMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}

	protocolDeviationID = args[0]

	subjectValBytes, err := getStateAndDecrypt(stub, t.bccspInst, transientMap, protocolDeviationID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + protocolDeviationID + "\"}"
		return shim.Error(jsonResp)
	}

	if subjectValBytes == nil {
		jsonResp := "{\"Error\":\"Nil value for " + protocolDeviationID + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(subjectValBytes)
}

func main() {
	factory.InitFactories(nil)

	err := shim.Start(&SimpleChaincode{factory.GetDefault()})
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
