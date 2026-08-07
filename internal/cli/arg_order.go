package cli

import (
	"fmt"
	"strings"
)

var longOptionArity = map[string]int{
	OptionExclude.Name: 0,
	OptionReset.Name:   0,
	OptionRunMode.Name: 1,
	OptionPartial.Name: 1,
	OptionOutput.Name:  1,
	OptionKill.Name:    0,
	OptionVersion.Name: 0,
	OptionHelp.Name:    0,
}

var shortOptionArity = map[byte]int{
	OptionExclude.Short[0]: 0,
	OptionReset.Short[0]:   0,
	OptionPartial.Short[0]: 1,
	OptionOutput.Short[0]:  1,
	OptionKill.Short[0]:    0,
	OptionVersion.Short[0]: 0,
	OptionHelp.Short[0]:    0,
}

func validateArgOrder(args []string) error {
	if len(args) == 0 {
		return nil
	}

	firstIsOption := isOptionToken(args[0])
	if !firstIsOption {
		return nil
	}

	for i := 0; i < len(args); i++ {
		tok := args[i]

		if !isOptionToken(tok) {
			return fmt.Errorf("when command starts with options, positional arguments are not allowed: %s", tok)
		}

		consumeNext, err := optionValueConsumption(tok)
		if err != nil {
			return err
		}
		if consumeNext {
			i++
		}
	}

	return nil
}

func isOptionToken(tok string) bool {
	return strings.HasPrefix(tok, "-") && tok != "-"
}

func optionValueConsumption(tok string) (bool, error) {
	if strings.HasPrefix(tok, "--") {
		name := strings.TrimPrefix(tok, "--")
		if eq := strings.IndexByte(name, '='); eq >= 0 {
			name = name[:eq]
		}
		arity, ok := longOptionArity[name]
		if !ok {
			return false, nil
		}
		if strings.Contains(tok, "=") {
			return false, nil
		}
		return arity == 1, nil
	}

	short := strings.TrimPrefix(tok, "-")
	if short == "" {
		return false, nil
	}

	for i := 0; i < len(short); i++ {
		arity, ok := shortOptionArity[short[i]]
		if !ok {
			continue
		}
		if arity == 0 {
			continue
		}

		if i < len(short)-1 {
			return false, nil
		}
		return true, nil
	}

	return false, nil
}
