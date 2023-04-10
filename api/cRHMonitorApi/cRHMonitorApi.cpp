#include <memory.h>
#include <assert.h>

#include "RHMonitorApi.h"
#include "cRHMonitorApi.h"

class C_API cRHMonitorApi : CRHMonitorSpi
{
public:
    cRHMonitorApi()
    {
        memset(&vtCallbacks, 0, sizeof(callback_t));

        createInstance();
    };
    cRHMonitorApi(callback_t *vt)
    {
        SetCallback(vt);

        createInstance();
    };

protected:
    ~cRHMonitorApi()
    {
        Release();
    }

private:
    CRHMonitorApi *pApi;
    callback_t vtCallbacks;

    void
    createInstance()
    {
        pApi = CRHMonitorApi::CreateRHMonitorApi();

        assert(pApi != NULL);

        pApi->RegisterSpi(this);
    }

public:
    void SetCallback(callback_t *vt)
    {
        memcpy(&vtCallbacks, vt, sizeof(callback_t));
    }

    void SetCbOnFrontConnected(CbOnFrontConnected handler)
    {
        vtCallbacks.cOnFrontConnected = handler;
    }

    void SetCbOnFrontDisconnected(CbOnFrontDisconnected handler)
    {
        vtCallbacks.cOnFrontDisconnected = handler;
    }

    void SetCbOnRspUserLogin(CbOnRspUserLogin handler)
    {
        vtCallbacks.cOnRspUserLogin = handler;
    }

    void SetCbOnRspUserLogout(CbOnRspUserLogout handler)
    {
        vtCallbacks.cOnRspUserLogout = handler;
    }

    void SetCbOnRspQryMonitorAccounts(CbOnRspQryMonitorAccounts handler)
    {
        vtCallbacks.cOnRspQryMonitorAccounts = handler;
    }

    void SetCbOnRspQryInvestorMoney(CbOnRspQryInvestorMoney handler)
    {
        vtCallbacks.cOnRspQryInvestorMoney = handler;
    }

    void SetCbOnRspQryInvestorPosition(CbOnRspQryInvestorPosition handler)
    {
        vtCallbacks.cOnRspQryInvestorPosition = handler;
    }

    void SetCbOnRspOffsetOrder(CbOnRspOffsetOrder handler)
    {
        vtCallbacks.cOnRspOffsetOrder = handler;
    }

    void SetCbOnRtnOrder(CbOnRtnOrder handler)
    {
        vtCallbacks.cOnRtnOrder = handler;
    }

    void SetCbOnRtnTrade(CbOnRtnTrade handler)
    {
        vtCallbacks.cOnRtnTrade = handler;
    }

    void SetCbOnRtnInvestorMoney(CbOnRtnInvestorMoney handler)
    {
        vtCallbacks.cOnRtnInvestorMoney = handler;
    }

    void SetCbOnRtnInvestorPosition(CbOnRtnInvestorPosition handler)
    {
        vtCallbacks.cOnRtnInvestorPosition = handler;
    }

    /// 初始化
    ///@remark 初始化运行环境,只有调用后,接口才开始工作
    void
    Init(const char *ip, unsigned int port)
    {
        pApi->Init(ip, port);
    };

    /// 账户登陆
    int ReqUserLogin(CRHMonitorReqUserLoginField *pUserLoginField, int nRequestID)
    {
        return pApi->ReqUserLogin(pUserLoginField, nRequestID);
    };

    // 账户登出
    int ReqUserLogout(CRHMonitorUserLogoutField *pUserLogoutField, int nRequestID)
    {
        return pApi->ReqUserLogout(pUserLogoutField, nRequestID);
    };

    // 查询所有管理的账户
    int ReqQryMonitorAccounts(CRHMonitorQryMonitorUser *pQryMonitorUser, int nRequestID)
    {
        return pApi->ReqQryMonitorAccounts(pQryMonitorUser, nRequestID);
    };

    /// 查询账户资金
    int ReqQryInvestorMoney(CRHMonitorQryInvestorMoneyField *pQryInvestorMoneyField, int nRequestID)
    {
        return pApi->ReqQryInvestorMoney(pQryInvestorMoneyField, nRequestID);
    };

    /// 查询账户持仓
    int ReqQryInvestorPosition(CRHMonitorQryInvestorPositionField *pQryInvestorPositionField, int nRequestID)
    {
        return pApi->ReqQryInvestorPosition(pQryInvestorPositionField, nRequestID);
    };

    // 给Server发送强平请求
    int ReqOffsetOrder(CRHMonitorOffsetOrderField *pMonitorOrderField, int nRequestID)
    {
        return pApi->ReqOffsetOrder(pMonitorOrderField, nRequestID);
    };

    // 订阅主动推送信息
    int ReqSubPushInfo(CRHMonitorSubPushInfo *pInfo, int nRequestID)
    {
        return pApi->ReqSubPushInfo(pInfo, nRequestID);
    };

    /// 当客户端与交易后台建立起通信连接时（还未登录前），该方法被调用。
    void OnFrontConnected()
    {
        if (!vtCallbacks.cOnFrontConnected)
            return;

        vtCallbacks.cOnFrontConnected(CRHMonitorInstance(this));
    };

    /// 当客户端与交易后台通信连接断开时，该方法被调用。当发生这个情况后，API会自动重新连接，客户端可不做处理。
    ///@param nReason 错误原因
    ///         0x1001 网络读失败
    ///         0x1002 网络写失败
    ///         0x2001 接收心跳超时
    ///         0x2002 发送心跳失败
    ///         0x2003 收到错误报文
    void OnFrontDisconnected(int nReason)
    {
        if (!vtCallbacks.cOnFrontDisconnected) return;

        vtCallbacks.cOnFrontDisconnected(CRHMonitorInstance(this), nReason);
    };

    /// 风控账户登陆响应
    void OnRspUserLogin(CRHMonitorRspUserLoginField *pRspUserLoginField, CRHRspInfoField *pRHRspInfoField, int nRequestID){
        if (!vtCallbacks.cOnRspUserLogin) return;

        vtCallbacks.cOnRspUserLogin(CRHMonitorInstance(this), pRspUserLoginField, pRHRspInfoField, nRequestID);
    };

    /// 风控账户登出响应
    void OnRspUserLogout(CRHMonitorUserLogoutField *pRspUserLogoutField, CRHRspInfoField *pRHRspInfoField, int nRequestID){
        if (!vtCallbacks.cOnRspUserLogout) return;

        vtCallbacks.cOnRspUserLogout(CRHMonitorInstance(this), pRspUserLogoutField, pRHRspInfoField, nRequestID);
    };

    // 查询监控账户响应
    void OnRspQryMonitorAccounts(CRHQryInvestorField *pRspMonitorUser, CRHRspInfoField *pRHRspInfoField, int nRequestID, bool isLast){
        if (!vtCallbacks.cOnRspQryMonitorAccounts) return;

        vtCallbacks.cOnRspQryMonitorAccounts(CRHMonitorInstance(this), pRspMonitorUser, pRHRspInfoField, nRequestID, isLast);
    };

    /// 查询账户资金响应
    void OnRspQryInvestorMoney(CRHTradingAccountField *pRHTradingAccountField, CRHRspInfoField *pRHRspInfoField, int nRequestID, bool isLast){
        if (!vtCallbacks.cOnRspQryInvestorMoney) return;

        vtCallbacks.cOnRspQryInvestorMoney(CRHMonitorInstance(this), pRHTradingAccountField, pRHRspInfoField, nRequestID, isLast);
    };

    /// 查询账户持仓信息响应
    void OnRspQryInvestorPosition(CRHMonitorPositionField *pRHMonitorPositionField, CRHRspInfoField *pRHRspInfoField, int nRequestID, bool isLast){
        if (!vtCallbacks.cOnRspQryInvestorPosition) return;

        vtCallbacks.cOnRspQryInvestorPosition(CRHMonitorInstance(this), pRHMonitorPositionField, pRHRspInfoField, nRequestID, isLast);
    };

    // 平仓指令发送失败时的响应
    void OnRspOffsetOrder(CRHMonitorOffsetOrderField *pMonitorOrderField, CRHRspInfoField *pRHRspInfoField, int nRequestID, bool isLast){
        if (!vtCallbacks.cOnRspOffsetOrder) return;

        vtCallbacks.cOnRspOffsetOrder(CRHMonitorInstance(this), pMonitorOrderField, pRHRspInfoField, nRequestID, isLast);
    };

    /// 报单通知
    void OnRtnOrder(CRHOrderField *pOrder){
        if (!vtCallbacks.cOnRtnOrder) return;

        vtCallbacks.cOnRtnOrder(CRHMonitorInstance(this), pOrder);
    };

    /// 成交通知
    void OnRtnTrade(CRHTradeField *pTrade){
        if (!vtCallbacks.cOnRtnTrade) return;

        vtCallbacks.cOnRtnTrade(CRHMonitorInstance(this), pTrade);
    };

    /// 账户资金发生变化回报
    void OnRtnInvestorMoney(CRHTradingAccountField *pRohonTradingAccountField){
        if (!vtCallbacks.cOnRtnInvestorMoney) return;

        vtCallbacks.cOnRtnInvestorMoney(CRHMonitorInstance(this), pRohonTradingAccountField);
    };

    /// 账户某合约持仓回报
    void OnRtnInvestorPosition(CRHMonitorPositionField *pRohonMonitorPositionField){
        if (!vtCallbacks.cOnRtnInvestorPosition) return;

        vtCallbacks.cOnRtnInvestorPosition(CRHMonitorInstance(this), pRohonMonitorPositionField);
    };

    void Release()
    {
        memset(&vtCallbacks, 0, sizeof(callback_t));

        if (NULL != pApi)
        {
            pApi->Release();
        }
    }
};

#ifdef __cplusplus
extern "C"
{
#endif

    C_API CRHMonitorInstance CreateRHMonitorApi()
    {
        cRHMonitorApi *instance = new cRHMonitorApi();

        return CRHMonitorInstance(instance);
    }

    C_API void Release(CRHMonitorInstance instance)
    {
        ((cRHMonitorApi *)instance)->Release();
    }

    C_API void Init(CRHMonitorInstance instance, const char *ip, unsigned int port)
    {
        ((cRHMonitorApi *)instance)->Init(ip, port);
    }

    C_API int ReqUserLogin(CRHMonitorInstance instance, struct CRHMonitorReqUserLoginField *pUserLoginField, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqUserLogin(pUserLoginField, nRequestID);
    }

    C_API int ReqUserLogout(CRHMonitorInstance instance, struct CRHMonitorUserLogoutField *pUserLogoutField, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqUserLogout(pUserLogoutField, nRequestID);
    }

    C_API int ReqQryMonitorAccounts(CRHMonitorInstance instance, struct CRHMonitorQryMonitorUser *pQryMonitorUser, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqQryMonitorAccounts(pQryMonitorUser, nRequestID);
    }

    C_API int ReqQryInvestorMoney(CRHMonitorInstance instance, struct CRHMonitorQryInvestorMoneyField *pQryInvestorMoneyField, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqQryInvestorMoney(pQryInvestorMoneyField, nRequestID);
    }

    C_API int ReqQryInvestorPosition(CRHMonitorInstance instance, struct CRHMonitorQryInvestorPositionField *pQryInvestorPositionField, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqQryInvestorPosition(pQryInvestorPositionField, nRequestID);
    }

    C_API int ReqOffsetOrder(CRHMonitorInstance instance, struct CRHMonitorOffsetOrderField *pMonitorOrderField, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqOffsetOrder(pMonitorOrderField, nRequestID);
    }

    C_API int ReqSubPushInfo(CRHMonitorInstance instance, struct CRHMonitorSubPushInfo *pInfo, int nRequestID)
    {
        return ((cRHMonitorApi *)instance)->ReqSubPushInfo(pInfo, nRequestID);
    }

    C_API void SetCallbacks(CRHMonitorInstance instance, callback_t *vt)
    {
        ((cRHMonitorApi *)instance)->SetCallback(vt);
    }

    C_API void SetCbOnFrontConnected(CRHMonitorInstance instance, CbOnFrontConnected handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnFrontConnected(handler);
    }

    C_API void SetCbOnFrontDisconnected(CRHMonitorInstance instance, CbOnFrontDisconnected handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnFrontDisconnected(handler);
    }

    C_API void SetCbOnRspUserLogin(CRHMonitorInstance instance, CbOnRspUserLogin handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspUserLogin(handler);
    }

    C_API void SetCbOnRspUserLogout(CRHMonitorInstance instance, CbOnRspUserLogout handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspUserLogout(handler);
    }

    C_API void SetCbOnRspQryMonitorAccounts(CRHMonitorInstance instance, CbOnRspQryMonitorAccounts handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspQryMonitorAccounts(handler);
    }

    C_API void SetCbOnRspQryInvestorMoney(CRHMonitorInstance instance, CbOnRspQryInvestorMoney handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspQryInvestorMoney(handler);
    }

    C_API void SetCbOnRspQryInvestorPosition(CRHMonitorInstance instance, CbOnRspQryInvestorPosition handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspQryInvestorPosition(handler);
    }

    C_API void SetCbOnRspOffsetOrder(CRHMonitorInstance instance, CbOnRspOffsetOrder handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRspOffsetOrder(handler);
    }

    C_API void SetCbOnRtnOrder(CRHMonitorInstance instance, CbOnRtnOrder handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRtnOrder(handler);
    }

    C_API void SetCbOnRtnTrade(CRHMonitorInstance instance, CbOnRtnTrade handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRtnTrade(handler);
    }

    C_API void SetCbOnRtnInvestorMoney(CRHMonitorInstance instance, CbOnRtnInvestorMoney handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRtnInvestorMoney(handler);
    }

    C_API void SetCbOnRtnInvestorPosition(CRHMonitorInstance instance, CbOnRtnInvestorPosition handler)
    {
        ((cRHMonitorApi *)instance)->SetCbOnRtnInvestorPosition(handler);
    }

#ifdef __cplusplus
}
#endif