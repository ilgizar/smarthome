package smarthome

// Usage configuration file structure

type UsageConfigPeriodStruct struct {
    Begin        int
    End          int
}

type UsageConfigConditionStruct struct {
    Date         []string
    DatePeriod   []UsageConfigPeriodStruct
    Time         []string
    TimePeriod   []UsageConfigPeriodStruct
    Weekday      []string
    Weekdays     []string
    Online       []string
    Offline      []string
    Quiet        bool
    Message      string
}

type UsageConfigActionStruct struct {
    Type         string
    Destination  string
    Value        string
}

type UsageConfigLimitedStruct struct {
    UsageConfigConditionStruct
    Overall      int
    Using        int
    Pause        int
    Action       []UsageConfigActionStruct
}

type UsageConfigOnlineStruct struct {
    UsageConfigConditionStruct
    After        int
    Before       int
    Pause        int
    Action       []UsageConfigActionStruct
}

type UsageConfigRuleStruct struct {
    Name         string
    Title        string
    Nodes        []string
    Enable       bool
    Enabled      bool
    Allowed      []UsageConfigConditionStruct
    Denied       []UsageConfigConditionStruct
    Limited      []UsageConfigLimitedStruct
    Online       []UsageConfigOnlineStruct
    Offline      []UsageConfigOnlineStruct
}

type UsageConfigStruct struct {
    Rule         []UsageConfigRuleStruct
}
