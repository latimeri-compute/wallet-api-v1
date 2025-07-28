package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/latimeri-compute/wallet-api-v1/internal/assert"
)

func TestShowWallet(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "существующий кошелёк",
			id:   "81a4c5c8-0085-45c1-9c44-d05912276715",
			want: `{"кошелёк":{"walletId":"81a4c5c8-0085-45c1-9c44-d05912276715","balance":1000}}`,
		},
		{
			name: "несуществующий кошелёк",
			id:   "01757132-f89b-48c6-aa57-1e8c9b2999d3",
			want: `{"ошибка":"запрашиваемый кошелёк не найден"}`,
		},
		{
			name: "неверный формат UUID",
			id:   "123123",
			want: `{"ошибка":"запрашиваемый кошелёк не найден"}`,
		},
		{
			name: "отсутвует id в ссылке",
			id:   "",
			want: "404 page not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			resp, err := http.Get(ts.URL + fmt.Sprintf("/api/v1/wallets/%s", test.id))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			b = bytes.TrimSpace(b)

			assert.Equal(t, string(b), test.want)
		})
	}
}

func TestChangeWalletBalance(t *testing.T) {
	app := newTestApplication(t)
	app.walletProcessorInput = app.StartWalletProcessor(app.walletModel)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name          string
		walletRequest walletRequestJSON
		want          string
	}{
		{
			name: "существующий кошелёк пополнение",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "deposit",
				Amount:        100,
			},
			want: `{"кошелёк":{"walletId":"81a4c5c8-0085-45c1-9c44-d05912276715","balance":1100}}`,
		},
		{
			name: "существующий кошелёк снятие",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "withdraw",
				Amount:        100,
			},
			want: `{"кошелёк":{"walletId":"81a4c5c8-0085-45c1-9c44-d05912276715","balance":900}}`,
		},
		{
			name: "несуществующий кошелёк",
			walletRequest: walletRequestJSON{
				ID:            "01757132-f89b-48c6-aa57-1e8c9b2999d3",
				OperationType: "deposit",
				Amount:        100,
			},
			want: `{"ошибка":"запрашиваемый кошелёк не найден"}`,
		},
		{
			name: "неверный формат UUID",
			walletRequest: walletRequestJSON{
				ID:            "123123",
				OperationType: "deposit",
				Amount:        100,
			},
			want: `{"ошибка":"запрашиваемый кошелёк не найден"}`,
		},
		{
			name: "снятие суммы больше счёта на кошельке",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "withdraw",
				Amount:        9000,
			},
			want: `{"ошибка":"недостаточный баланс кошелька для совершения операции"}`,
		},
		{
			name: "amount равен нулю",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "deposit",
			},
			want: `{"ошибка":{"amount":"значение поля не может быть пустым"}}`,
		},
		{
			name: "amount меньше нуля",
			walletRequest: walletRequestJSON{
				ID:            "81a4c5c8-0085-45c1-9c44-d05912276715",
				OperationType: "withdraw",
				Amount:        -46,
			},
			want: `{"ошибка":{"amount":"значение поля должно быть больше нуля"}}`,
		},
		{
			name: "пустой json",
			want: `{"ошибка":{"amount":"значение поля не может быть пустым","operationType":"значение поля не может быть пустым","walletId":"значение поля не может быть пустым"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonValue, err := json.Marshal(test.walletRequest)
			if err != nil {
				t.Fatal(err)
			}

			b := sendRequest(t, ts, jsonValue)

			assert.Equal(t, string(b), test.want)
		})
	}

	testsBrokenJson := []struct {
		name string
		json string
		want string
	}{
		{
			name: "отсутствует тело запроса",
			want: `{"ошибка":"тело запроса не должено быть пустым"}`,
		},
		{
			name: "незнакомые поля",
			json: `{"none": ":)"}`,
			want: `{"ошибка":"тело содержит неизвестное поле \"none\""}`,
		},
		{
			name: "незакрытая фигурная скобка",
			json: `{"walletId": "81a4c5c8-0085-45c1-9c44-d05912276715",`,
			want: `{"ошибка":"тело запроса содержит неправильно составленный JSON"}`,
		},
	}
	for _, test := range testsBrokenJson {
		t.Run(test.name, func(t *testing.T) {
			reqJSON := []byte(test.json)
			b := sendRequest(t, ts, reqJSON)

			assert.Equal(t, string(b), test.want)
		})
	}
}
