#pragma once
#ifndef C_MONITOR_API_H
#define C_MONITOR_API_H

#if defined(ISLIB) && defined(WIN32)
#ifdef CRHMONITORAPI_EXPORTS
#define C_API __declspec(dllexport)
#else
#define C_API __declspec(dllimport)
#endif
#else
#ifdef CRHMONITORAPI_EXPORTS
#define C_API __attribute__((visibility("default")))
#else
#define C_API __attribute__((visibility("default")))
#endif
#endif

#ifndef _WIN32
#define __cdecl
#endif
#define APPWINAPI __cdecl

#ifdef __cplusplus
#ifndef NULL
#define NULL 0
#endif
extern "C"
{
#else
#define NULL ((void *)0)
#endif

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <memory.h>

#include "RHUserApiDataType.h"
#include "RHUserApiStruct.h"

    // C API 实例
    typedef uintptr_t CRHMonitorInstance;

    /// 当客户端与交易后台建立起通信连接时（还未登录前），该方法被调用。
    typedef void(APPWINAPI *CbOnFrontConnected)(CRHMonitorInstance instance);
    C_API void SetCbOnFrontConnected(CRHMonitorInstance instance, CbOnFrontConnected handler);

    /// 当客户端与交易后台通信连接断开时，该方法被调用。当发生这个情况后，API会自动重新连接，客户端可不做处理。
    ///@param nReason 错误原因
    ///         0x1001 网络读失败
    ///         0x1002 网络写失败
    ///         0x2001 接收心跳超时
    ///         0x2002 发送心跳失败
    ///         0x2003 收到错误报文
    typedef void(APPWINAPI *CbOnFrontDisconnected)(CRHMonitorInstance instance, int nReason);
    C_API void SetCbOnFrontDisconnected(CRHMonitorInstance instance, CbOnFrontDisconnected handler);

    /// 风控账户登陆响应
    typedef void(APPWINAPI *CbOnRspUserLogin)(
        CRHMonitorInstance instance,
        struct CRHMonitorRspUserLoginField *pRspUserLoginField,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID);
    C_API void SetCbOnRspUserLogin(CRHMonitorInstance instance, CbOnRspUserLogin handler);

    /// 风控账户登出响应
    typedef void(APPWINAPI *CbOnRspUserLogout)(
        CRHMonitorInstance instance,
        struct CRHMonitorUserLogoutField *pRspUserLogoutField,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID);
    C_API void SetCbOnRspUserLogout(CRHMonitorInstance instance, CbOnRspUserLogout handler);

    /// 查询监控账户响应
    typedef void(APPWINAPI *CbOnRspQryMonitorAccounts)(
        CRHMonitorInstance instance,
        struct CRHQryInvestorField *pRspMonitorUser,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID, bool isLast);
    C_API void SetCbOnRspQryMonitorAccounts(CRHMonitorInstance instance, CbOnRspQryMonitorAccounts handler);

    /// 查询账户资金响应
    typedef void(APPWINAPI *CbOnRspQryInvestorMoney)(
        CRHMonitorInstance instance,
        struct CRHTradingAccountField *pRHTradingAccountField,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID, bool isLast);
    C_API void SetCbOnRspQryInvestorMoney(CRHMonitorInstance instance, CbOnRspQryInvestorMoney handler);

    /// 查询账户持仓信息响应
    typedef void(APPWINAPI *CbOnRspQryInvestorPosition)(
        CRHMonitorInstance instance,
        struct CRHMonitorPositionField *pRHMonitorPositionField,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID, bool isLast);
    C_API void SetCbOnRspQryInvestorPosition(CRHMonitorInstance instance, CbOnRspQryInvestorPosition handler);

    /// 平仓指令发送失败时的响应
    typedef void(APPWINAPI *CbOnRspOffsetOrder)(
        CRHMonitorInstance instance,
        struct CRHMonitorOffsetOrderField *pMonitorOrderField,
        struct CRHRspInfoField *pRHRspInfoField,
        int nRequestID, bool isLast);
    C_API void SetCbOnRspOffsetOrder(CRHMonitorInstance instance, CbOnRspOffsetOrder handler);

    /// 报单通知
    typedef void(APPWINAPI *CbOnRtnOrder)(CRHMonitorInstance instance, struct CRHOrderField *pOrder);
    C_API void SetCbOnRtnOrder(CRHMonitorInstance instance, CbOnRtnOrder handler);

    /// 成交通知
    typedef void(APPWINAPI *CbOnRtnTrade)(CRHMonitorInstance instance, struct CRHTradeField *pTrade);
    C_API void SetCbOnRtnTrade(CRHMonitorInstance instance, CbOnRtnTrade handler);

    /// 账户资金发生变化回报
    typedef void(APPWINAPI *CbOnRtnInvestorMoney)(CRHMonitorInstance instance, struct CRHTradingAccountField *pRohonTradingAccountField);
    C_API void SetCbOnRtnInvestorMoney(CRHMonitorInstance instance, CbOnRtnInvestorMoney handler);

    /// 账户某合约持仓回报
    typedef void(APPWINAPI *CbOnRtnInvestorPosition)(CRHMonitorInstance instance, struct CRHMonitorPositionField *pRohonMonitorPositionField);
    C_API void SetCbOnRtnInvestorPosition(CRHMonitorInstance instance, CbOnRtnInvestorPosition handler);

    typedef struct CbVirtualTable
    {
        CbOnFrontConnected cOnFrontConnected;
        CbOnFrontDisconnected cOnFrontDisconnected;
        CbOnRspUserLogin cOnRspUserLogin;
        CbOnRspUserLogout cOnRspUserLogout;
        CbOnRspQryMonitorAccounts cOnRspQryMonitorAccounts;
        CbOnRspQryInvestorMoney cOnRspQryInvestorMoney;
        CbOnRspQryInvestorPosition cOnRspQryInvestorPosition;
        CbOnRspOffsetOrder cOnRspOffsetOrder;
        CbOnRtnOrder cOnRtnOrder;
        CbOnRtnTrade cOnRtnTrade;
        CbOnRtnInvestorMoney cOnRtnInvestorMoney;
        CbOnRtnInvestorPosition cOnRtnInvestorPosition;
    } callback_t;

    C_API void SetCallbacks(CRHMonitorInstance instance, callback_t *vt);

    /// 创建MonitorApi
    C_API CRHMonitorInstance CreateRHMonitorApi();
    /// 删除接口对象本身
    ///@remark 不再使用本接口对象时,调用该函数删除接口对象
    C_API void Release(CRHMonitorInstance instance);

    /// 初始化
    ///@remark 初始化运行环境,只有调用后,接口才开始工作
    C_API void Init(CRHMonitorInstance instance, const char *ip, unsigned int port);

    /// 账户登陆
    C_API int ReqUserLogin(CRHMonitorInstance instance, struct CRHMonitorReqUserLoginField *pUserLoginField, int nRequestID);

    // 账户登出
    C_API int ReqUserLogout(CRHMonitorInstance instance, struct CRHMonitorUserLogoutField *pUserLogoutField, int nRequestID);

    // 查询所有管理的账户
    C_API int ReqQryMonitorAccounts(CRHMonitorInstance instance, struct CRHMonitorQryMonitorUser *pQryMonitorUser, int nRequestID);

    /// 查询账户资金
    C_API int ReqQryInvestorMoney(CRHMonitorInstance instance, struct CRHMonitorQryInvestorMoneyField *pQryInvestorMoneyField, int nRequestID);

    /// 查询账户持仓
    C_API int ReqQryInvestorPosition(CRHMonitorInstance instance, struct CRHMonitorQryInvestorPositionField *pQryInvestorPositionField, int nRequestID);

    // 给Server发送强平请求
    C_API int ReqOffsetOrder(CRHMonitorInstance instance, struct CRHMonitorOffsetOrderField *pMonitorOrderField, int nRequestID);

    // 订阅主动推送信息
    C_API int ReqSubPushInfo(CRHMonitorInstance instance, struct CRHMonitorSubPushInfo *pInfo, int nRequestID);

#ifdef __cplusplus
}
#endif

#endif