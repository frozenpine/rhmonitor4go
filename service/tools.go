package service

import rohon "github.com/frozenpine/rhmonitor4go"

func RohonCaculateDynamicBalance(
	preBalance, deposit, withdraw, closeProfit, openProfit, commission float64,
) float64 {
	return preBalance + deposit - withdraw + closeProfit + openProfit - commission
}

func RohonCompareAccount(pre, curr *rohon.Account) bool {
	preBalance := RohonCaculateDynamicBalance(
		pre.PreBalance, pre.Deposit, pre.Withdraw,
		pre.CloseProfit, pre.PositionProfit, pre.Commission,
	)

	currBalance := RohonCaculateDynamicBalance(
		curr.PreBalance, curr.Deposit, curr.Withdraw,
		curr.CloseProfit, curr.PositionProfit, curr.Commission,
	)

	return preBalance == currBalance
}
