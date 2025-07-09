package report

import (
	"fmt"
	"time"

	"firefly-task/pkg/interfaces"
)

func createTestDriftResults() map[string]*interfaces.DriftResult {
	results := make(map[string]*interfaces.DriftResult)

	// Resource with drift
	webServer1 := &interfaces.DriftResult{
		ResourceID:   "i-1234567890abcdef0",
		ResourceType: "aws_instance",
		IsDrifted:      true,
		Severity:       interfaces.SeverityMedium,
		DetectionTime:  time.Now(),
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "instance_type",
				ExpectedValue: "t2.micro",
				ActualValue:   "t2.small",
				Severity:      interfaces.SeverityMedium,
			},
		},
	}
	results["aws_instance.web-server-1"] = webServer1

	// Resource with critical drift
	webServer2 := &interfaces.DriftResult{
		ResourceID:   "i-abcdef1234567890",
		ResourceType: "aws_instance",
		IsDrifted:      true,
		Severity:       interfaces.SeverityCritical,
		DetectionTime:  time.Now(),
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "security_groups",
				ExpectedValue: "sg-1234",
				ActualValue:   "sg-5678",
				Severity:      interfaces.SeverityCritical,
			},
		},
	}
	results["aws_instance.web-server-2"] = webServer2

	// Resource without drift
	db := &interfaces.DriftResult{
		ResourceID:   "i-fedcba9876543210",
		ResourceType: "aws_db_instance",
		IsDrifted:      false,
		Severity:       interfaces.SeverityLow,
		DetectionTime:  time.Now(),
		DriftDetails: []*interfaces.DriftDetail{},
	}
	results["aws_db_instance.database"] = db

	// Resource with another drift
	lb := &interfaces.DriftResult{
		ResourceID:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188",
		ResourceType: "aws_lb",
		IsDrifted:      true,
		Severity:       interfaces.SeverityHigh,
		DetectionTime:  time.Now(),
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "publicly_accessible",
				ExpectedValue: "true",
				ActualValue:   "false",
				Severity:      interfaces.SeverityHigh,
			},
		},
	}
	results["aws_lb.main"] = lb

	return results
}

func createLargeDriftResults(count int) map[string]*interfaces.DriftResult {
	results := make(map[string]*interfaces.DriftResult)
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("aws_instance.web-server-%d", i)
		instanceId := fmt.Sprintf("i-abcdef%d", i)
		result := &interfaces.DriftResult{
			ResourceID:   instanceId,
			ResourceType: "aws_instance",
			IsDrifted:      true,
			Severity:       interfaces.SeverityMedium,
			DetectionTime:  time.Now(),
			DriftDetails: []*interfaces.DriftDetail{
				{
					Attribute:     "instance_type",
					ExpectedValue: "t2.micro",
					ActualValue:   "t2.small",
					Severity:      interfaces.SeverityMedium,
				},
			},
		}
		results[id] = result
	}
	return results
}