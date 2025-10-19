package session

type TransformRule struct {
	Forward  func(string) string
	Backward func(string) string
}

func NewTransformRule(forward func(string) string, backward func(string) string) TransformRule {
	return TransformRule{
		Forward:  forward,
		Backward: backward,
	}
}

type Transformer struct {
	rules []TransformRule
}

func NewTransformer() *Transformer {
	return &Transformer{
		rules: make([]TransformRule, 0),
	}
}

func (tf *Transformer) WithRule(rules ...TransformRule) *Transformer {
	tf.rules = append(tf.rules, rules...)

	return tf
}

func (tf *Transformer) Transform(in string) string {
	for _, rule := range tf.rules {
		in = rule.Forward(in)
	}

	return in
}

func (tf *Transformer) Revert(in string) string {
	for i := len(tf.rules) - 1; i >= 0; i-- {
		in = tf.rules[i].Backward(in)
	}

	return in
}
