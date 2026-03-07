package contestio

import (
	"bytes"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

type parserFunc func([]byte) (any, error)

func test_parseInt(t *testing.T, parser func(any) parserFunc) {
	tests := []struct {
		name     string
		input    string
		expected any // ожидаемое значение для каждого типа (будем проверять кастомно)
		wantErr  bool
		errType  error // конкретная ошибка (опционально)
	}{
		// Базовые корректные значения
		{"zero", "0", 0, false, nil},
		{"positive", "123", 123, false, nil},
		{"negative_signed", "-456", -456, false, nil},

		// Граничные значения для int8
		{"int8_zero", "0", int8(0), false, nil},
		{"int8_max", "127", int8(127), false, nil},
		{"int8_min", "-128", int8(-128), false, nil},
		{"int8_overflow_plus", "128", int8(0), true, strconv.ErrRange},
		{"int8_overflow_minus", "-129", int8(0), true, strconv.ErrRange},

		// Граничные значения для int16
		{"int16_zero", "0", int16(0), false, nil},
		{"int16_max", "32767", int16(32767), false, nil},
		{"int16_min", "-32768", int16(-32768), false, nil},
		{"int16_overflow_plus", "32768", int16(0), true, strconv.ErrRange},
		{"int16_overflow_minus", "-32769", int16(0), true, strconv.ErrRange},

		// Граничные значения для int32
		{"int32_zero", "0", int32(0), false, nil},
		{"int32_max", "2147483647", int32(2147483647), false, nil},
		{"int32_min", "-2147483648", int32(-2147483648), false, nil},
		{"int32_overflow_plus", "2147483648", int32(0), true, strconv.ErrRange},
		{"int32_overflow_minus", "-2147483649", int32(0), true, strconv.ErrRange},

		// Граничные значения для int64
		{"int64_zero", "0", int64(0), false, nil},
		{"int64_max", "9223372036854775807", int64(9223372036854775807), false, nil},
		{"int64_min", "-9223372036854775808", int64(-9223372036854775808), false, nil},
		{"int64_overflow_plus", "9223372036854775808", int64(0), true, strconv.ErrRange},
		{"int64_overflow_minus", "-9223372036854775809", int64(0), true, strconv.ErrRange},

		// Беззнаковые
		{"uint8_zero", "0", uint8(0), false, nil},
		{"uint8_max", "255", uint8(255), false, nil},
		{"uint8_overflow", "256", uint8(0), true, strconv.ErrRange},
		{"uint16_zero", "0", uint16(0), false, nil},
		{"uint16_max", "65535", uint16(65535), false, nil},
		{"uint16_overflow", "65536", uint16(0), true, strconv.ErrRange},
		{"uint32_zero", "0", uint32(0), false, nil},
		{"uint32_max", "4294967295", uint32(4294967295), false, nil},
		{"uint32_overflow", "4294967296", uint32(0), true, strconv.ErrRange},
		{"uint64_zero", "0", uint64(0), false, nil},
		{"uint64_max", "18446744073709551615", uint64(18446744073709551615), false, nil},
		{"uint64_overflow", "18446744073709551616", uint64(0), true, strconv.ErrRange},

		// Отрицательные для беззнаковых – должны быть ошибкой
		{"uint_negative", "-1", uint(0), true, strconv.ErrSyntax},
		{"uint8_negative", "-1", uint8(0), true, strconv.ErrSyntax},
		{"uint16_negative", "-1", uint16(0), true, strconv.ErrSyntax},
		{"uint32_negative", "-1", uint32(0), true, strconv.ErrSyntax},
		{"uint64_negative", "-1", uint64(0), true, strconv.ErrSyntax},

		// Невалидный ввод
		{"empty", "", int(0), true, strconv.ErrSyntax},
		{"too long", strings.Repeat("9", 42), int(0), true, strconv.ErrRange},
		{"only_sign", "+", int(0), true, strconv.ErrSyntax},
		{"invalid", "abc", int(0), true, strconv.ErrSyntax},
		{"leading_space", " 123", int(0), true, strconv.ErrSyntax},
		{"trailing_space", "123 ", int(0), true, strconv.ErrSyntax},
		{"plus_sign_signed", "+123", int64(123), false, nil},               // ParseInt и Atoi принимает '+'
		{"plus_sign_unsigned", "+123", uint64(0), true, strconv.ErrSyntax}, // ParseUint не принимает '+'
		{"184467440737095516159", "184467440737095516159", uint64(0), true, strconv.ErrRange},

		// Двойные/перекрёстные знаки
		{"double_minus", "--123", int(0), true, strconv.ErrSyntax},
		{"double_plus", "++123", int(0), true, strconv.ErrSyntax},
		{"mixed_signs_1", "+-123", int(0), true, strconv.ErrSyntax},
		{"mixed_signs_2", "-+123", int(0), true, strconv.ErrSyntax},

		// Ведущие нули (должны работать)
		{"leading_zeros", "000123", int64(123), false, nil},
		{"leading_zeros_overflow", "00000000000000000000128", int8(0), true, strconv.ErrRange},
		{"only_zeros", "0000", int64(0), false, nil},

		// Числа с десятичной точкой / экспонентой (частая ошибка ввода)
		{"float_input", "123.45", int64(0), true, strconv.ErrSyntax},
		{"scientific_notation", "1e5", int64(0), true, strconv.ErrSyntax},
		{"scientific_capital", "1E+3", int64(0), true, strconv.ErrSyntax},

		// Шестнадцатеричные/восьмеричные префиксы (должны отвергаться)
		{"hex_prefix", "0x10", int64(0), true, strconv.ErrSyntax},
		{"octal_prefix", "010", int64(10), false, nil}, // если не запрещаете явно — 010 = 10
		{"hex_letters", "12a3", int64(0), true, strconv.ErrSyntax},

		// Валидные цифры + мусор после
		{"valid_then_invalid", "123abc", int64(0), true, strconv.ErrSyntax},
		{"invalid_then_valid", "abc123", int64(0), true, strconv.ErrSyntax},
		{"digit_then_sign", "123-", int64(0), true, strconv.ErrSyntax},

		// Null-байты и не-ASCII (защита от некорректных буферов)
		{"null_byte_prefix", "\x00123", int64(0), true, strconv.ErrSyntax},
		{"null_byte_suffix", "123\x00", int64(0), true, strconv.ErrSyntax},
		{"unicode_digit", "¹²³", int64(0), true, strconv.ErrSyntax}, // superscript, не валидно
		{"emoji", "12🔥3", int64(0), true, strconv.ErrSyntax},

		// Граничные 20-значные числа (проверка оптимизации)
		{"exact_20_digits_uint64_max", "18446744073709551615", uint64(18446744073709551615), false, nil},
		{"exact_20_digits_overflow", "18446744073709551616", uint64(0), true, strconv.ErrRange},
		{"21_digits", "100000000000000000000", uint64(0), true, strconv.ErrRange},
		{"negative_20_digits", "-9223372036854775808", int64(-9223372036854775808), false, nil},
		{"negative_20_digits_overflow", "-9223372036854775809", int64(0), true, strconv.ErrRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Вызываем parseInt с соответствующим типом
			got, err := parser(tt.expected)([]byte(tt.input))

			// Проверка ошибки
			if tt.wantErr {
				if err == nil {
					t.Logf("got %[1]v (%[1]T), expected %[2]v (%[2]T)", got, tt.expected)
					t.Errorf("expected error, got nil")
				} else {
					// Проверяем конкретный тип ошибки, если задан
					if tt.errType != nil {
						if !errors.Is(err, tt.errType) {
							t.Errorf("expected error type %v, got %v", tt.errType, err)
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			// Если ошибки не ожидается, сравниваем значение
			if !tt.wantErr {
				var success bool
				switch want := tt.expected.(type) {
				case int:
					v, ok := got.(int)
					success = ok && v == want
				case int8:
					v, ok := got.(int8)
					success = ok && v == want
				case int16:
					v, ok := got.(int16)
					success = ok && v == want
				case int32:
					v, ok := got.(int32)
					success = ok && v == want
				case int64:
					v, ok := got.(int64)
					success = ok && v == want
				case uint:
					v, ok := got.(uint)
					success = ok && v == want
				case uint8:
					v, ok := got.(uint8)
					success = ok && v == want
				case uint16:
					v, ok := got.(uint16)
					success = ok && v == want
				case uint32:
					v, ok := got.(uint32)
					success = ok && v == want
				case uint64:
					v, ok := got.(uint64)
					success = ok && v == want
				}

				if !success {
					t.Errorf("got %[1]v (%[1]T), expected %[2]v (%[2]T)", got, tt.expected)
				}
			}
		})
	}
}

func Test_ParseIntStd(t *testing.T) {
	test_parseInt(t, func(expected any) parserFunc {
		switch expected.(type) {
		case int:
			return func(b []byte) (any, error) { return parseIntStd[int](b) }
		case int8:
			return func(b []byte) (any, error) { return parseIntStd[int8](b) }
		case int16:
			return func(b []byte) (any, error) { return parseIntStd[int16](b) }
		case int32:
			return func(b []byte) (any, error) { return parseIntStd[int32](b) }
		case int64:
			return func(b []byte) (any, error) { return parseIntStd[int64](b) }
		case uint:
			return func(b []byte) (any, error) { return parseIntStd[uint](b) }
		case uint8:
			return func(b []byte) (any, error) { return parseIntStd[uint8](b) }
		case uint16:
			return func(b []byte) (any, error) { return parseIntStd[uint16](b) }
		case uint32:
			return func(b []byte) (any, error) { return parseIntStd[uint32](b) }
		case uint64:
			return func(b []byte) (any, error) { return parseIntStd[uint64](b) }
		default:
			panic("unsupported type")
		}
	})
}

func Test_parseIntBase(t *testing.T) {
	test_parseInt(t, func(expected any) parserFunc {
		switch expected.(type) {
		case int:
			return func(b []byte) (any, error) { return parseIntBase[int](b) }
		case int8:
			return func(b []byte) (any, error) { return parseIntBase[int8](b) }
		case int16:
			return func(b []byte) (any, error) { return parseIntBase[int16](b) }
		case int32:
			return func(b []byte) (any, error) { return parseIntBase[int32](b) }
		case int64:
			return func(b []byte) (any, error) { return parseIntBase[int64](b) }
		case uint:
			return func(b []byte) (any, error) { return parseIntBase[uint](b) }
		case uint8:
			return func(b []byte) (any, error) { return parseIntBase[uint8](b) }
		case uint16:
			return func(b []byte) (any, error) { return parseIntBase[uint16](b) }
		case uint32:
			return func(b []byte) (any, error) { return parseIntBase[uint32](b) }
		case uint64:
			return func(b []byte) (any, error) { return parseIntBase[uint64](b) }
		default:
			panic("unsupported type")
		}
	})
}

func Test_parseIntFast(t *testing.T) {
	test_parseInt(t, func(expected any) parserFunc {
		switch expected.(type) {
		case int:
			return func(b []byte) (any, error) { return parseIntFast[int](b) }
		case int8:
			return func(b []byte) (any, error) { return parseIntFast[int8](b) }
		case int16:
			return func(b []byte) (any, error) { return parseIntFast[int16](b) }
		case int32:
			return func(b []byte) (any, error) { return parseIntFast[int32](b) }
		case int64:
			return func(b []byte) (any, error) { return parseIntFast[int64](b) }
		case uint:
			return func(b []byte) (any, error) { return parseIntFast[uint](b) }
		case uint8:
			return func(b []byte) (any, error) { return parseIntFast[uint8](b) }
		case uint16:
			return func(b []byte) (any, error) { return parseIntFast[uint16](b) }
		case uint32:
			return func(b []byte) (any, error) { return parseIntFast[uint32](b) }
		case uint64:
			return func(b []byte) (any, error) { return parseIntFast[uint64](b) }
		default:
			panic("unsupported type")
		}
	})
}

func fuzz_parseInt(f *testing.F, parseInt func([]byte) (int64, error)) {

	f.Fuzz(func(t *testing.T, input []byte) {
		// Сравниваем поведение с strconv для валидных десятичных строк
		// (игнорируем случаи, где strconv ведёт себя иначе, например '+' для unsigned)

		// Быстрая проверка: только цифры и опциональный знак в начале
		trimmed := bytes.TrimLeft(input, "+-")
		for _, b := range trimmed {
			if b < '0' || b > '9' {
				// Наш парсер должен вернуть ошибку
				_, err := parseInt(input)
				if err == nil {
					t.Errorf("expected error for invalid input %q", input)
				}
				return
			}
		}

		// Если строка валидна — сверяем результат со strconv
		// (для signed типов)
		if len(input) > 0 && (input[0] == '-' || input[0] == '+') {
			// strconv.ParseInt принимает '+', ваш парсер для signed — тоже
			want, err := strconv.ParseInt(string(input), 10, 64)
			got, gotErr := parseInt(input)

			if (err != nil) != (gotErr != nil) {
				t.Errorf("input %q: error mismatch: strconv=%v, parseInt=%v", input, err, gotErr)
			}
			if err == nil && got != want {
				t.Errorf("input %q: got %d, want %d", input, got, want)
			}
		}
	})
}

var parseIntRes int64

func Benchmark_parseInt(b *testing.B) {
	b.StopTimer()
	N := 1 << 20 // 1M
	rand := rand.New(rand.NewSource(1))
	nums := generateInts[int](rand, N)
	input := makeIntsInput(rand, nums, 1)
	tokens := bytes.Fields(input)

	b.Run("parseIntStd", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, token := range tokens {
				parseIntRes, _ = parseIntStd[int64](token)
			}
		}
	})

	b.Run("parseIntBase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, token := range tokens {
				parseIntRes, _ = parseIntBase[int64](token)
			}
		}
	})

	b.Run("parseIntFast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, token := range tokens {
				parseIntRes, _ = parseIntFast[int64](token)
			}
		}
	})
}

func Benchmark_parseInt_boundaries(b *testing.B) {
	b.StopTimer()

	cases := [][]byte{
		[]byte("0"),
		[]byte("127"), []byte("128"), // int8 boundary
		[]byte("32767"), []byte("32768"), // int16
		[]byte("2147483647"), []byte("2147483648"), // int32
		[]byte("9223372036854775807"), []byte("9223372036854775808"), // int64
		[]byte("18446744073709551615"), []byte("18446744073709551616"), // uint64
		bytes.Repeat([]byte("9"), 19), // 19 digits
		bytes.Repeat([]byte("9"), 20), // 20 digits - hot path
		bytes.Repeat([]byte("9"), 21), // 21 digits - early reject
	}

	for _, c := range cases {
		b.Run(string(c)+"_std", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseIntRes, _ = parseIntStd[int64](c)
			}
		})
		b.Run(string(c), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseIntRes, _ = parseInt[int64](c)
			}
		})
	}
}
