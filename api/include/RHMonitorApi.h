/////////////////////////////////////////////////////////////////////////
///@system �ں��ڻ�����ƽ̨
///@company �Ϻ��ں���Ϣ�������޹�˾
///@file RHMonitorApi.h
///@brief �����˿ͻ��˽ӿ�
///20180910 create by Haosc
///
/////////////////////////////////////////////////////////////////////////

#if !defined(RH_RHMONITORAPI_H)
#define RH_RHMONITORAPI_H

#if _MSC_VER > 1000
#pragma once
#endif // _MSC_VER > 1000

#include "RHUserApiStruct.h"

#if defined(ISLIB) && defined(WIN32)
	#ifdef LIB_RHMONITOR_API_EXPORT
		#define RHMONITOR_API_EXPORT_NEW __declspec(dllexport)
	#else
		#define RHMONITOR_API_EXPORT_NEW __declspec(dllimport)
	#endif
#else
	#ifdef LIB_RHMONITOR_API_EXPORT
		#define RHMONITOR_API_EXPORT_NEW __attribute__((visibility("default")))
	#else
		#define RHMONITOR_API_EXPORT_NEW
	#endif
#endif

class CRHMonitorSpi
{
public:
	///���ͻ����뽻�׺�̨������ͨ������ʱ����δ��¼ǰ�����÷��������á�
	virtual void OnFrontConnected(){};
	
	///���ͻ����뽻�׺�̨ͨ�����ӶϿ�ʱ���÷��������á���������������API���Զ��������ӣ��ͻ��˿ɲ�������
	///@param nReason ����ԭ��
	///        0x1001 �����ʧ��
	///        0x1002 ����дʧ��
	///        0x2001 ����������ʱ
	///        0x2002 ��������ʧ��
	///        0x2003 �յ�������
	virtual void OnFrontDisconnected(int nReason){};

	///����˻���½��Ӧ
	virtual void OnRspUserLogin(CRHMonitorRspUserLoginField* pRspUserLoginField, CRHRspInfoField* pRHRspInfoField, int nRequestID) {};

	///����˻��ǳ���Ӧ
	virtual void OnRspUserLogout(CRHMonitorUserLogoutField* pRspUserLoginField, CRHRspInfoField* pRHRspInfoField, int nRequestID) {};

	//��ѯ����˻���Ӧ
	virtual void OnRspQryMonitorAccounts(CRHQryInvestorField* pRspMonitorUser,CRHRspInfoField* pRHRspInfoField, int nRequestID,bool isLast) {};

	///��ѯ�˻��ʽ���Ӧ 
	virtual void OnRspQryInvestorMoney(CRHTradingAccountField* pRHTradingAccountField, CRHRspInfoField* pRHRspInfoField, int nRequestID,bool isLast) {};

	///��ѯ�˻��ֲ���Ϣ��Ӧ
	virtual void OnRspQryInvestorPosition(CRHMonitorPositionField* pRHMonitorPositionField, CRHRspInfoField* pRHRspInfoField, int nRequestID,bool isLast) {};

	//ƽ��ָ���ʧ��ʱ����Ӧ
	virtual void OnRspOffsetOrder(CRHMonitorOffsetOrderField* pMonitorOrderField,CRHRspInfoField* pRHRspInfoField, int nRequestID,bool isLast){};

	///����֪ͨ
	virtual void OnRtnOrder(CRHOrderField *pOrder) {};
	
	///�ɽ�֪ͨ
	virtual void OnRtnTrade(CRHTradeField *pTrade) {};

	///�˻��ʽ����仯�ر�
	virtual void OnRtnInvestorMoney(CRHTradingAccountField* pRHTradingAccountField) {};

	///�˻�ĳ��Լ�ֲֻر�
	virtual void OnRtnInvestorPosition(CRHMonitorPositionField* pRHMonitorPositionField) {};

	///��ѯ������Ӧ
	virtual void OnRspQryOrder(CRHOrderField *pOrder, CRHRspInfoField *pRspInfo, int nRequestID, bool bIsLast) {};

	//�����ѯ���гɽ��ر���Ӧ
	virtual void OnRspQryTrade(CRHTradeField *pTrade, CRHRspInfoField *pRspInfo, int nRequestID, bool bIsLast) {};

	///��ѯ��Լ��Ϣ��Ӧ
	virtual void OnRspQryInstrument(CRHMonitorInstrumentField* pRHMonitorInstrumentField,CRHRspInfoField* pRHRspInfoField, int nRequestID,bool isLast) {};
};

class RHMONITOR_API_EXPORT_NEW CRHMonitorApi
{
public:
	///����MonitorApi
	static CRHMonitorApi *CreateRHMonitorApi();
	///ɾ���ӿڶ�����
	///@remark ����ʹ�ñ��ӿڶ���ʱ,���øú���ɾ���ӿڶ���
	virtual void Release() = 0;
	
	///��ʼ��
	///@remark ��ʼ�����л���,ֻ�е��ú�,�ӿڲſ�ʼ����
	virtual void Init(const char * ip,unsigned int port) = 0;

	///ע��ص��ӿ�
	///@param pSpi �����Իص��ӿ����ʵ��
	virtual void RegisterSpi(CRHMonitorSpi *pSpi) = 0;

	///�˻���½
	virtual int ReqUserLogin(CRHMonitorReqUserLoginField* pUserLoginField, int nRequestID) = 0;

	//�˻��ǳ�
	virtual int ReqUserLogout(CRHMonitorUserLogoutField* pUserLogoutField, int nRequestID) = 0;

	//��ѯ���й�����˻�
	virtual int ReqQryMonitorAccounts(CRHMonitorQryMonitorUser* pQryMonitorUser, int nRequestID) = 0;

	///��ѯ�˻��ʽ�
	virtual int ReqQryInvestorMoney(CRHMonitorQryInvestorMoneyField* pQryInvestorMoneyField, int nRequestID) = 0;

	///��ѯ�˻��ֲ�
	virtual int ReqQryInvestorPosition(CRHMonitorQryInvestorPositionField* pQryInvestorPositionField, int nRequestID) = 0;

	//��Server����ǿƽ����
	virtual int ReqOffsetOrder(CRHMonitorOffsetOrderField* pMonitorOrderField, int nRequestID) = 0;

	//��������������Ϣ
	virtual int ReqSubPushInfo(CRHMonitorSubPushInfo *pInfo, int nRequestID) = 0;

	//��ѯ�����˻�����ί�м�¼����ǰֻ֧�ֲ�ѯ�����˻���BrokerID�ÿ�
	virtual int ReqQryOrder(CRHQryOrderField* pQryOrder, int nRequestID) = 0;

	//��ѯ�����˻����ճɽ���¼����ǰֻ֧�ֲ�ѯ�����˻���BrokerID�ÿ�
	virtual int ReqQryTrade(CRHQryTradeField* pQryTrade, int nRequestID) = 0;

	///��ѯ��Լ��Ϣ,Ĭ�Ϻ�Լ����д�����к�Լ
	virtual int ReqQryInstrument(CRHMonitorInstrumentField* pQryInstrumentField, int nRequestID) = 0;
};

#endif