package oc2oscal

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/oscalkit/pkg/oc2oscal/masonry"
	"github.com/docker/oscalkit/pkg/oscal/constants"
	"github.com/docker/oscalkit/types/oscal"
	ssp "github.com/docker/oscalkit/types/oscal/system_security_plan"
	"github.com/docker/oscalkit/types/oscal/validation_root"
	"github.com/opencontrol/compliance-masonry/pkg/lib/common"
)

func Convert(repoUri, outputDirectory string) error {
	workspace, err := masonry.Open(repoUri)
	if err != nil {
		return err
	}

	_, err = os.Stat(outputDirectory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(outputDirectory, 0755)
	}
	if err != nil {
		return err
	}

	var metadata ssp.Metadata
	metadata.Title = ssp.Title("FedRAMP System Security Plan (SSP)")
	metadata.LastModified = validation_root.LastModified(time.Now().Format(constants.FormatDatetimeTz))
	metadata.Version = validation_root.Version("0.0.1")
	metadata.OscalVersion = validation_root.OscalVersion(constants.LatestOscalVersion)

	for _, component := range workspace.GetAllComponents() {
		err = convertComponent(component, metadata, outputDirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertComponent(component common.Component, metadata ssp.Metadata, outputDirectory string) error {
	var plan ssp.SystemSecurityPlan
	plan.Id = "TODO"
	plan.Metadata = &metadata
	plan.ImportProfile = &ssp.ImportProfile{
		Href: "https://raw.githubusercontent.com/usnistgov/OSCAL/master/content/fedramp.gov/xml/FedRAMP_MODERATE-baseline_profile.xml",
	}
	plan.SystemCharacteristics = convertSystemCharacteristics(component)
	return writeSSP(plan, outputDirectory+"/"+component.GetKey()+".xml")
}

func convertSystemCharacteristics(component common.Component) *ssp.SystemCharacteristics {
	var syschar ssp.SystemCharacteristics
	syschar.SystemIds = []ssp.SystemId{
		ssp.SystemId{
			IdentifierType: "https://fedramp.gov",
			Value:          "F00000000",
		},
	}
	syschar.SystemName = ssp.SystemName(component.GetName())
	syschar.SystemNameShort = ssp.SystemNameShort(component.GetKey())
	syschar.Description = &ssp.Description{
		Raw: "<p>Automatically generated OSCAL SSP from OpenControl guidance for " + component.GetName() + "</p>",
	}
	syschar.SecuritySensitivityLevel = ssp.SecuritySensitivityLevel("low")
	syschar.SystemInformation = staticSystemInformation()
	syschar.SecurityImpactLevel = &ssp.SecurityImpactLevel{
		SecurityObjectiveConfidentiality: ssp.SecurityObjectiveConfidentiality("fips-199-moderate"),
		SecurityObjectiveIntegrity:       ssp.SecurityObjectiveIntegrity("fips-199-moderate"),
		SecurityObjectiveAvailability:    ssp.SecurityObjectiveAvailability("fips-199-moderate"),
	}
	syschar.Status = &ssp.Status{
		State: "operational",
	}
	syschar.AuthorizationBoundary = &ssp.AuthorizationBoundary{
		Description: &ssp.Description{
			Raw: "<p>A holistic, top-level explanation of the FedRAMP authorization boundary.</p>",
		},
	}
	return &syschar
}

func staticSystemInformation() *ssp.SystemInformation {
	var sysinf ssp.SystemInformation
	sysinf.InformationTypes = []ssp.InformationType{
		ssp.InformationType{
			Name: "Information Type Name",
			Description: &ssp.Description{
				Raw: "<p>This item is useless nevertheless required.</p>",
			},
			ConfidentialityImpact: &ssp.ConfidentialityImpact{
				Base: "fips-199-moderate",
			},
			IntegrityImpact: &ssp.IntegrityImpact{
				Base: "fips-199-moderate",
			},
			AvailabilityImpact: &ssp.AvailabilityImpact{
				Base: "fips-199-moderate",
			},
		},
	}
	return &sysinf
}

func writeSSP(plan ssp.SystemSecurityPlan, outputFile string) error {
	destFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("Error opening output file %s: %s", outputFile, err)
	}
	defer destFile.Close()

	output := oscal.OSCAL{SystemSecurityPlan: &plan}
	return output.XML(destFile, true)
}
