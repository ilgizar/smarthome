package smarthome

// Usage configuration file structure

type UsageConfigPeriodStruct struct {
    Begin        int
    End          int
}

type UsageConfigActionStruct struct {
    Type         string
    Destination  string
    Value        string
    Disable      bool
    After        int
    Before       int
    Pause        int

    Enabled      bool
}

type UsageConfigConditionStruct struct {
    Date         []string
    Time         []string
    Weekday      []string
    Online       []string
    Offline      []string
    Quiet        bool
    Message      string
    Action       []UsageConfigActionStruct

    DatePeriod   []UsageConfigPeriodStruct
    TimePeriod   []UsageConfigPeriodStruct
    Weekdays     []string
}

type UsageConfigLimitedStruct struct {
    UsageConfigConditionStruct
    Overall      int
    Using        int
    Pause        int
}

type UsageConfigRuleStruct struct {
    Name         string
    Title        string
    Nodes        []string
    Disable      bool
    Allowed      []UsageConfigConditionStruct
    Denied       []UsageConfigConditionStruct
    Limited      []UsageConfigLimitedStruct
    Online       []UsageConfigConditionStruct
    Offline      []UsageConfigConditionStruct

    Enabled      bool
}

type UsageConfigStruct struct {
    Rule         []UsageConfigRuleStruct
}
