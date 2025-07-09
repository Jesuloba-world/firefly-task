package drift

import (
	"testing"

	"firefly-task/pkg/interfaces"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConcreteDriftDetector(t *testing.T) {
	logger := logrus.New()
	detector := NewConcreteDriftDetector(logger)
	assert.NotNil(t, detector)
}

func TestNewConcreteDriftComparator(t *testing.T) {
	logger := logrus.New()
	comparator := NewConcreteDriftComparator(logger)
	assert.NotNil(t, comparator)
}

func TestNewConcreteDriftAnalyzer(t *testing.T) {
	logger := logrus.New()
	analyzer := NewConcreteDriftAnalyzer(logger)
	assert.NotNil(t, analyzer)
}

func TestConcreteDriftAnalyzer_AnalyzeDriftSeverity(t *testing.T) {
	analyzer := NewConcreteDriftAnalyzer(nil)

	t.Run("with drift details", func(t *testing.T) {
		driftResult := &interfaces.DriftResult{
			DriftDetails: []*interfaces.DriftDetail{{Attribute: "instance_type"}},
		}
		severity := analyzer.AnalyzeDriftSeverity(driftResult)
		assert.Equal(t, interfaces.SeverityHigh, severity)
	})

	t.Run("without drift details", func(t *testing.T) {
		driftResult := &interfaces.DriftResult{}
		severity := analyzer.AnalyzeDriftSeverity(driftResult)
		assert.Equal(t, interfaces.SeverityNone, severity)
	})
}

func TestConcreteDriftDetector_DetectDrift(t *testing.T) {
	detector := NewConcreteDriftDetector(nil)
	actual := &interfaces.EC2Instance{}
	expected := &interfaces.TerraformConfig{}

	result, err := detector.DetectDrift(actual, expected, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestConcreteDriftDetector_DetectMultipleDrift(t *testing.T) {
	detector := NewConcreteDriftDetector(nil)
	actualResources := map[string]*interfaces.EC2Instance{
		"resource1": {},
	}
	expectedConfigs := map[string]*interfaces.TerraformConfig{
		"resource1": {},
	}

	results, err := detector.DetectMultipleDrift(actualResources, expectedConfigs, nil)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 1)
}

func TestConcreteDriftDetector_ValidateConfiguration(t *testing.T) {
	detector := NewConcreteDriftDetector(nil)

	t.Run("valid config", func(t *testing.T) {
		config := &interfaces.TerraformConfig{}
		err := detector.ValidateConfiguration(config)
		assert.NoError(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		err := detector.ValidateConfiguration(nil)
		assert.Error(t, err)
	})
}

func TestConcreteDriftComparator_CompareAttribute(t *testing.T) {
	comparator := NewConcreteDriftComparator(nil)
	_, err := comparator.CompareAttribute("instance_type", "t2.micro", "t2.small")
	assert.NoError(t, err)
}

func TestConcreteDriftComparator_CompareAttributes(t *testing.T) {
	comparator := NewConcreteDriftComparator(nil)
	actual := map[string]interface{}{}
	expected := map[string]interface{}{}
	_, err := comparator.CompareAttributes(actual, expected, nil)
	assert.NoError(t, err)
}

func TestConcreteDriftComparator_GetSupportedAttributes(t *testing.T) {
	comparator := NewConcreteDriftComparator(nil)
	attributes := comparator.GetSupportedAttributes()
	assert.Contains(t, attributes, "instance_type")
	assert.Contains(t, attributes, "tags")
}

func TestConcreteDriftAnalyzer_GroupDriftByType(t *testing.T) {
	analyzer := NewConcreteDriftAnalyzer(nil)
	results := map[string]*interfaces.DriftResult{}
	grouped := analyzer.GroupDriftByType(results)
	assert.Nil(t, grouped)
}

func TestConcreteDriftAnalyzer_FilterDriftByAttribute(t *testing.T) {
	analyzer := NewConcreteDriftAnalyzer(nil)
	results := map[string]*interfaces.DriftResult{}
	filtered := analyzer.FilterDriftByAttribute(results, nil)
	assert.Nil(t, filtered)
}

func TestConcreteDriftAnalyzer_CalculateDriftStatistics(t *testing.T) {
	analyzer := NewConcreteDriftAnalyzer(nil)
	results := map[string]*interfaces.DriftResult{
		"resource1": {},
	}
	stats := analyzer.CalculateDriftStatistics(results)
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.TotalResources)
}
