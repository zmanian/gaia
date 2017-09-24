package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecimalFromString(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    Decimal
		wantErr bool
	}{
		{"test 0.0", args{"0.0"}, NewDecimal(0, 0), false},
		{"test 0.01", args{"0.01"}, NewDecimal(1, -2), false},
		{"test 1.0", args{"1"}, NewDecimal(1, 0), false},
		{"test 500.0", args{"500"}, NewDecimal(5, 2), false},
		{"test -10.0", args{"-10"}, NewDecimal(-1, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDecimalFromString(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDecimalFromString(%v) error = %v, wantErr %v", tt.args.value, err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("NewDecimalFromString(%v) = %v, want %v", tt.args.value, got, tt.want)
			}
		})
	}
}

func TestDecimal_Round(t *testing.T) {
	type fields struct {
		Value int64
		Exp   int32
	}
	type args struct {
		places int32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Decimal
	}{
		{"0.06 round 1", fields{6, -2}, args{1}, NewDecimal(1, -1)},
		{"0.04 round 1", fields{4, -2}, args{1}, NewDecimal(0, 0)},
		{"0.05 round 1", fields{5, -2}, args{1}, NewDecimal(0, 0)},
		{"0.15 round 1", fields{15, -2}, args{1}, NewDecimal(2, -1)},
		{"0.25 round 1", fields{25, -2}, args{1}, NewDecimal(2, -1)},
		{"1.25 round 1", fields{115, -2}, args{1}, NewDecimal(12, -1)},
		//{"0.06 round 10", fields{6, -2}, args{10}, NewDecimal(6, -2)},
		//{"0.06 round 18", fields{6, -2}, args{18}, NewDecimal(6, -2)},
		//{"500 round 16", fields{5, 2}, args{16}, NewDecimal(5, 2)},
		//{"500 round 17", fields{5, 2}, args{17}, NewDecimal(5, 2)},
		//{"500 round 18", fields{5, 2}, args{18}, NewDecimal(5, 2)},
		//{"500 round 19", fields{5, 2}, args{19}, NewDecimal(5, 2)},
		//{"500 round 20", fields{5, 2}, args{20}, NewDecimal(5, 2)},
		//{"50 round 16", fields{5, 1}, args{16}, NewDecimal(5, 1)},
		//{"50 round 17", fields{5, 1}, args{17}, NewDecimal(5, 1)},
		//{"50 round 18", fields{5, 1}, args{18}, NewDecimal(5, 1)},
		//{"50 round 19", fields{5, 1}, args{19}, NewDecimal(5, 1)},
		//{"50 round 20", fields{5, 1}, args{20}, NewDecimal(5, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Decimal{
				Value: tt.fields.Value,
				Exp:   tt.fields.Exp,
			}
			assert.True(t, d.Round(tt.args.places).Equal(tt.want), tt.name)
		})
	}
}

//func TestDecimal_String(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//tests := []struct {
//name   string
//fields fields
//want   string
//}{
//{"test 0.0", fields{0, 0}, "0.0"},
//{"test 0.01", fields{1, -2}, "0.01"},
//{"test 1.0", fields{1, 0}, "1.0"},
//{"test 500.0", fields{5, 2}, "500.0"},
//{"test -10.0", fields{-1, 2}, "-10.0"},
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.String(); got != tt.want {
//t.Errorf("Decimal.String() = %v, want %v", got, tt.want)
//}
//})
//}
//}

//func TestDecimal_IntPart(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//tests := []struct {
//name   string
//fields fields
//want   int64
//}{
//{"test 0.0", fields{0, 0}, int64(0)},
//{"test 0.01", fields{1, -2}, int64(0)},
//{"test 1.0", fields{1, 0}, int64(1)},
//{"test 500.0", fields{5, 2}, int64(500)},
//{"test -10.0", fields{-1, 1}, int64(-10)},
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.IntPart(); got != tt.want {
//t.Errorf("Decimal.IntPart() = %v, want %v", got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Add(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   Decimal
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Add(tt.args.d2); !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Decimal.Add(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Sub(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   Decimal
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Sub(tt.args.d2); !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Decimal.Sub(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Negative(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//tests := []struct {
//name   string
//fields fields
//want   Decimal
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Negative(); !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Decimal.Negative() = %v, want %v", got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Mul(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   Decimal
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Mul(tt.args.d2); !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Decimal.Mul(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Div(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   Decimal
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Div(tt.args.d2); !reflect.DeepEqual(got, tt.want) {
//t.Errorf("Decimal.Div(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_Equal(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   bool
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.Equal(tt.args.d2); got != tt.want {
//t.Errorf("Decimal.Equal(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_GT(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   bool
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.GT(tt.args.d2); got != tt.want {
//t.Errorf("Decimal.GT(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_GTE(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   bool
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.GTE(tt.args.d2); got != tt.want {
//t.Errorf("Decimal.GTE(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_LT(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   bool
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.LT(tt.args.d2); got != tt.want {
//t.Errorf("Decimal.LT(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}

//func TestDecimal_LTE(t *testing.T) {
//type fields struct {
//Value int64
//Exp   int32
//}
//type args struct {
//d2 Decimal
//}
//tests := []struct {
//name   string
//fields fields
//args   args
//want   bool
//}{
//// TODO: Add test cases.
//}
//for _, tt := range tests {
//t.Run(tt.name, func(t *testing.T) {
//d := Decimal{
//Value: tt.fields.Value,
//Exp:   tt.fields.Exp,
//}
//if got := d.LTE(tt.args.d2); got != tt.want {
//t.Errorf("Decimal.LTE(%v) = %v, want %v", tt.args.d2, got, tt.want)
//}
//})
//}
//}
