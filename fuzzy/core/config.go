package core

// Параметры по умолчанию.
const (
	DefaultAlphaLevels = 50
	DefaultGridPoints  = 1000
	DefaultEps         = 1e-9
)

// Config — параметры численных методов.
// Нулевые поля заменяются значениями по умолчанию,
type Config struct {
	AlphaLevels int     // число шагов по α при арифметике и интегрировании
	GridPoints  int     // число точек сетки по x (для метрик по μ(x))
	Eps         float64 // допуск при сравнении чисел
}

func DefaultConfig() Config {
	return Config{AlphaLevels: DefaultAlphaLevels, GridPoints: DefaultGridPoints, Eps: DefaultEps}
}

func (c Config) WithDefaults() Config {
	if c.AlphaLevels < 1 {
		c.AlphaLevels = DefaultAlphaLevels
	}
	if c.GridPoints < 1 {
		c.GridPoints = DefaultGridPoints
	}
	if c.Eps <= 0 {
		c.Eps = DefaultEps
	}
	return c
}
