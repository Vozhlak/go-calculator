package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type CommandAction int

const (
	NoAction CommandAction = iota
	Continue
	Break
)

type Calculator struct {
	history []string
}

func clearConsole() {
	cmd := exec.Command("clear") // Для Linux/MacOS
	if _, err := exec.LookPath("cmd.exe"); err == nil {
		cmd = exec.Command("cmd", "/c", "cls") // Для Windows
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return
	}
}

func tokenize(expression string) ([]string, error) {
	regex := regexp.MustCompile(`\d+(\.\d+)?|[+\-*/()]`)

	tokens := regex.FindAllString(expression, -1)
	if tokens == nil {
		return nil, fmt.Errorf("не удалось токенизировать выражение: %s", expression)
	}

	return tokens, nil
}

func (c *Calculator) calcFunc(values []float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return values[0] + values[1], nil
	case "-":
		return values[0] - values[1], nil
	case "*":
		return values[0] * values[1], nil
	case "/":
		if values[1] == 0 {
			return 0, errors.New("деление на ноль невозможно")
		}
		return values[0] / values[1], nil
	default:
		return 0, nil
	}
}

func (c *Calculator) handleCommand(input string) CommandAction {
	switch input {
	case "help", "h":
		clearConsole()
		c.getInfoTheCommand()
		return Continue
	case "exit", "q":
		fmt.Println("Выход из программы")
		return Break
	case "show", "s":
		clearConsole()
		c.showHistory()
		return Continue
	default:
		return NoAction
	}
}

func (c *Calculator) getInfoTheCommand() {
	fmt.Println("Доступные команды:")
	fmt.Println("help / h - Показать справку")
	fmt.Println("exit / q - Выход из программы")
	fmt.Println("show / s - Показать историю вычислений")
}

func (c *Calculator) addHistory(input string, result float64) {
	modifyInput := strings.ReplaceAll(input, " ", "")
	expression := fmt.Sprintf("%s = %.2f", modifyInput, result)
	c.history = append(c.history, expression)
}

func (c *Calculator) showHistory() {
	if len(c.history) == 0 {
		fmt.Println("История пуста")
		return
	}

	fmt.Println("История вычислений:")
	for _, record := range c.history {
		fmt.Println(record)
	}
}

func (c *Calculator) parseToRPN(expression string) ([]string, error) {
	tokens, errTokenize := tokenize(expression)

	if errTokenize != nil {
		return nil, errTokenize
	}

	var output []string
	var stack []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return nil, errors.New("некорректное выражение: нет соответствующей открывающей скобки")
			}
			stack = stack[:len(stack)-1]
		} else if precedence[token] > 0 {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			return nil, fmt.Errorf("некорректный токен: %s", token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, errors.New("некорректное выражение: нет соответствующей закрывающей скобки")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func (c *Calculator) evaluateRPN(rpn []string) (float64, error) {
	var stack []float64

	for i, token := range rpn {
		if value, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, value)
		} else {
			if len(stack) < 2 {
				return 0, fmt.Errorf("ошибка: недостаточно операндов для выполнения операции '%s' на позиции %d", token, i+1)
			}

			a := stack[len(stack)-1]
			b := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			result, err := c.calcFunc([]float64{b, a}, token)
			if err != nil {
				return 0, fmt.Errorf("ошибка при вычислении операции '%s': %v", token, err)
			}

			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return 0, errors.New("ошибка: выражение некорректно, остались лишние или отсутствующие операнды")
	}

	return stack[0], nil
}

func (c *Calculator) run() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Введите help или h для вывода информации существующих команд")
	fmt.Println()
	fmt.Println("Введите текст (или exit или q для выхода):")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		action := c.handleCommand(input)

		if action == Break {
			break
		} else if action == Continue {
			continue
		}

		rpn, errParseToRPN := c.parseToRPN(input)
		if errParseToRPN != nil {
			fmt.Printf("Ошибка: %s\n", errParseToRPN)
			continue
		}

		result, errEvaluateRPN := c.evaluateRPN(rpn)
		if errEvaluateRPN != nil {
			fmt.Printf("Ошибка: %s\n", errEvaluateRPN)
			continue
		}

		fmt.Printf("Результат: %.2f\n", result)
		c.addHistory(input, result)
	}
}

func main() {
	fmt.Println("Happy coding!!!")
	fmt.Println()

	calc := new(Calculator)
	calc.run()
}
