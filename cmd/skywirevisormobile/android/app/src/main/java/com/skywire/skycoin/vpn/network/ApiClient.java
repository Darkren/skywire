package com.skywire.skycoin.vpn.network;

import com.skywire.skycoin.vpn.network.models.VpnServerModel;

import java.util.List;

import io.reactivex.rxjava3.core.Observable;
import retrofit2.Response;
import retrofit2.Retrofit;
import retrofit2.adapter.rxjava3.RxJava3CallAdapterFactory;
import retrofit2.converter.gson.GsonConverterFactory;
import retrofit2.http.GET;
import retrofit2.http.Query;
import retrofit2.http.Url;

public class ApiClient {

    private interface ApiInterface {
        @GET("services")
        Observable<Response<List<VpnServerModel>>> getVpnServers(@Query("type") String type);

        @GET
        Observable<Response<Void>> checkConnection(@Url String url);
    }

    public static final String BASE_URL = "https://service.discovery.skycoin.com/api/";

    private static final Retrofit retrofit = new Retrofit.Builder()
        .baseUrl(BASE_URL)
        .addConverterFactory(GsonConverterFactory.create())
        .addCallAdapterFactory(RxJava3CallAdapterFactory.createSynchronous())
        .build();

    private static final ApiInterface apiService = retrofit.create(ApiInterface.class);

    public static Observable<Response<List<VpnServerModel>>> getVpnServers() {
        return apiService.getVpnServers("vpn");
    }

    public static Observable<Response<Void>> checkConnection(String url) {
        return apiService.checkConnection(url);
    }
}
