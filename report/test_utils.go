package report

import (
	"fmt"

	"firefly-task/drift"
)

func createTestDriftResults() map[string]*drift.DriftResult {
	results := make(map[string]*drift.DriftResult)

	// Resource with drift
	webServer1 := drift.NewDriftResult("aws_instance.web-server-1", "i-1234567890abcdef0")
	webServer1.ResourceType = "aws_instance"
	webServer1.AddDifference(drift.AttributeDifference{
		AttributeName: "instance_type",
		ExpectedValue: "t2.micro",
		ActualValue:   "t2.small",
		Severity:      drift.SeverityMedium,
	})
	results["aws_instance.web-server-1"] = webServer1

	// Resource with critical drift
	webServer2 := drift.NewDriftResult("aws_instance.web-server-2", "i-abcdef1234567890")
	webServer2.ResourceType = "aws_instance"
	webServer2.AddDifference(drift.AttributeDifference{
		AttributeName: "security_groups",
		ExpectedValue: "sg-1234",
		ActualValue:   "sg-5678",
		Severity:      drift.SeverityCritical,
	})
	results["aws_instance.web-server-2"] = webServer2

	// Resource without drift
	db := drift.NewDriftResult("aws_db_instance.database", "i-fedcba9876543210")
	db.ResourceType = "aws_db_instance"
	results["aws_db_instance.database"] = db

	// Resource with another drift
	lb := drift.NewDriftResult("aws_lb.main", "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188")
	lb.ResourceType = "aws_lb"
	lb.AddDifference(drift.AttributeDifference{
		AttributeName: "publicly_accessible",
		ExpectedValue: "true",
		ActualValue:   "false",
		Severity:      drift.SeverityHigh,
	})
	results["aws_lb.main"] = lb

	return results
}

func createLargeDriftResults(count int) map[string]*drift.DriftResult {
	results := make(map[string]*drift.DriftResult)
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("aws_instance.web-server-%d", i)
		instanceId := fmt.Sprintf("i-abcdef%d", i)
		result := drift.NewDriftResult(id, instanceId)
		result.ResourceType = "aws_instance"
		result.AddDifference(drift.AttributeDifference{
			AttributeName: "instance_type",
			ExpectedValue: "t2.micro",
			ActualValue:   "t2.small",
			Severity:      drift.SeverityMedium,
		})
		results[id] = result
	}
	return results
}