package com.skywire.skycoin.vpn.activities.servers;

import android.content.SharedPreferences;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.preference.PreferenceManager;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.google.gson.Gson;
import com.skywire.skycoin.vpn.App;
import com.skywire.skycoin.vpn.R;
import com.skywire.skycoin.vpn.activities.index.IndexPageAdapter;
import com.skywire.skycoin.vpn.controls.Tab;
import com.skywire.skycoin.vpn.extensible.ClickEvent;
import com.skywire.skycoin.vpn.helpers.HelperFunctions;
import com.skywire.skycoin.vpn.objects.LocalServerData;
import com.skywire.skycoin.vpn.objects.ServerFlags;
import com.skywire.skycoin.vpn.objects.ServerRatings;
import com.skywire.skycoin.vpn.vpn.VPNServersPersistentData;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.Date;

import io.reactivex.rxjava3.core.Observable;
import io.reactivex.rxjava3.disposables.Disposable;

public class ServersActivity extends Fragment implements VpnServersAdapter.VpnServerSelectedListener, ClickEvent {
    public static String ADDRESS_DATA_PARAM = "address";
    private static final String ACTIVE_TAB_KEY = "activeTab";

    private Tab tabPublic;
    private Tab tabHistory;
    private Tab tabFavorites;
    private Tab tabBlocked;
    private RecyclerView recycler;
    private ProgressBar loadingAnimation;
    private TextView textNoResults;

    private IndexPageAdapter.RequestTabListener requestTabListener;
    private ServerLists listType = ServerLists.Public;
    private VpnServersAdapter adapter;
    private SharedPreferences settings = PreferenceManager.getDefaultSharedPreferences(App.getContext());

    private Disposable serverSubscription;

    @Nullable
    @Override
    public View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container, @Nullable Bundle savedInstanceState) {
        super.onCreateView(inflater, container, savedInstanceState);

        return inflater.inflate(R.layout.activity_server_list, container, true);
    }

    @Override
    public void onViewCreated(View view, Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);

        tabPublic = view.findViewById(R.id.tabPublic);
        tabHistory = view.findViewById(R.id.tabHistory);
        tabFavorites = view.findViewById(R.id.tabFavorites);
        tabBlocked = view.findViewById(R.id.tabBlocked);
        recycler = view.findViewById(R.id.recycler);
        loadingAnimation = view.findViewById(R.id.loadingAnimation);
        textNoResults = view.findViewById(R.id.textNoResults);

        tabPublic.setClickEventListener(this);
        tabHistory.setClickEventListener(this);
        tabFavorites.setClickEventListener(this);
        tabBlocked.setClickEventListener(this);

        LinearLayoutManager layoutManager = new LinearLayoutManager(getContext());
        recycler.setLayoutManager(layoutManager);
        // This could be useful in the future.
        // recycler.setHasFixedSize(true);

        // This code retrieves the data from the server and populates the list with the recovered
        // data, but is not used right now as the server is returning empty arrays.
        // requestData()

        // Initialize the recycler.
        adapter = new VpnServersAdapter(getContext());
        adapter.setData(new ArrayList<>(), listType);
        adapter.setVpnSelectedEventListener(this);
        recycler.setAdapter(adapter);

        Gson gson = new Gson();
        String savedlistType = settings.getString(ACTIVE_TAB_KEY, null);
        if (savedlistType != null) {
            listType = gson.fromJson(savedlistType, ServerLists.class);
        }

        showCorrectList();
    }

    public void setRequestTabListener(IndexPageAdapter.RequestTabListener listener) {
        requestTabListener = listener;
    }

    @Override
    public void onClick(View view) {
        if (view.getId() == R.id.tabPublic) {
            listType = ServerLists.Public;
        } else if (view.getId() == R.id.tabHistory) {
            listType = ServerLists.History;
        } else if (view.getId() == R.id.tabFavorites) {
            listType = ServerLists.Favorites;
        } else if (view.getId() == R.id.tabBlocked) {
            listType = ServerLists.Blocked;
        }

        Gson gson = new Gson();
        String listTypeString = gson.toJson(listType);
        settings.edit()
            .putString(ACTIVE_TAB_KEY, listTypeString)
            .apply();

        showCorrectList();
    }

    private void showCorrectList() {
        tabPublic.changeState(false);
        tabHistory.changeState(false);
        tabFavorites.changeState(false);
        tabBlocked.changeState(false);

        if (listType == ServerLists.Public) {
            tabPublic.changeState(true);
            // Use test data, for now.
            showTestServers();
        } else {
            if (listType == ServerLists.History) {
                tabHistory.changeState(true);
            } else if (listType == ServerLists.Favorites) {
                tabFavorites.changeState(true);
            } else if (listType == ServerLists.Blocked) {
                tabBlocked.changeState(true);
            }

            requestLocalData();
        }
    }

    private void requestData() {
        if (serverSubscription != null) {
            serverSubscription.dispose();
        }

        /*
        serverSubscription = ApiClient.getVpnServers()
            .subscribeOn(Schedulers.io())
            .observeOn(AndroidSchedulers.mainThread())
            .subscribe(response -> {
                VpnServersAdapter adapter = new VpnServersAdapter(this, response.body());
                adapter.setVpnSelectedEventListener(this);
                recycler.setAdapter(adapter);
            }, err -> {
                this.requestData();
            });
        */
    }

    private void requestLocalData() {
        if (serverSubscription != null) {
            serverSubscription.dispose();
        }

        textNoResults.setVisibility(View.GONE);
        recycler.setVisibility(View.GONE);
        loadingAnimation.setVisibility(View.VISIBLE);

        Observable<ArrayList<LocalServerData>> request;
        if (listType == ServerLists.History) {
            request = VPNServersPersistentData.getInstance().history();
        } else if (listType == ServerLists.Favorites) {
            request = VPNServersPersistentData.getInstance().favorites();
        } else {
            request = VPNServersPersistentData.getInstance().blocked();
        }

        serverSubscription = request.subscribe(response -> {
            ArrayList<VpnServerForList> list = new ArrayList<>();

            for (LocalServerData server : response) {
                list.add(convertLocalServerData(server));
            }

            sortList(list);
            adapter.setData(list, listType);

            recycler.setVisibility(View.VISIBLE);
            loadingAnimation.setVisibility(View.GONE);

            if (list.size() == 0) {
                textNoResults.setVisibility(View.VISIBLE);
            }
        });
    }

    public static VpnServerForList convertLocalServerData(LocalServerData server) {
        VpnServerForList converted = new VpnServerForList();

        converted.countryCode = server.countryCode;
        converted.name = server.name;
        converted.customName = server.customName;
        converted.location = server.location;
        converted.pk = server.pk;
        converted.note = server.note;
        converted.personalNote = server.personalNote;
        converted.lastUsed = server.lastUsed;
        converted.inHistory = server.inHistory;
        converted.flag = server.flag;
        converted.enteredManually = server.enteredManually;
        converted.usedWithPassword = server.usedWithPassword;

        return converted;
    }

    @Override
    public void onResume() {
        super.onResume();
        //HelperFunctions.closeActivityIfServiceRunning(this);
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();

        if (serverSubscription != null) {
            serverSubscription.dispose();
        }
    }

    @Override
    public void onVpnServerSelected(VpnServerForList selectedServer) {
        start(VPNServersPersistentData.getInstance().processFromList(selectedServer));
        /*
        if (HelperFunctions.closeActivityIfServiceRunning(this)) {
            return;
        }

        Intent resultIntent = new Intent();
        resultIntent.putExtra(ADDRESS_DATA_PARAM, selectedServer.pk);
        setResult(RESULT_OK, resultIntent);
        finish();
        */
    }

    @Override
    public void onManualEntered(LocalServerData server) {
        start(server);
        /*
        if (HelperFunctions.closeActivityIfServiceRunning(this)) {
            return;
        }

        Intent resultIntent = new Intent();
        resultIntent.putExtra(ADDRESS_DATA_PARAM, server.pk);
        setResult(RESULT_OK, resultIntent);
        finish();
        */
    }

    private void start(LocalServerData server) {
        boolean starting = HelperFunctions.prepareAndStartVpn(getActivity(), server);

        if (starting) {
            if (requestTabListener != null) {
                requestTabListener.onOpenStatusRequested();
            }
        }
    }

    private void showTestServers() {
        ArrayList<VpnServerForList> servers = new ArrayList<>();

        VpnServerForList testServer = new VpnServerForList();
        testServer.lastUsed = new Date();
        testServer.countryCode = "au";
        testServer.name = "Server name";
        testServer.location = "Melbourne";
        testServer.pk = "024ec47420176680816e0406250e7156465e4531f5b26057c9f6297bb0303558c7";
        testServer.congestion = 20;
        testServer.congestionRating = ServerRatings.Gold;
        testServer.latency = 123;
        testServer.latencyRating = ServerRatings.Gold;
        testServer.hops = 3;
        testServer.note = "Note";
        servers.add(testServer);

        testServer = new VpnServerForList();
        testServer.lastUsed = new Date();
        testServer.countryCode = "br";
        testServer.name = "Test server 14";
        testServer.location = "Rio de Janeiro";
        testServer.pk = "034ec47420176680816e0406250e7156465e4531f5b26057c9f6297bb0303558c7";
        testServer.congestion = 20;
        testServer.congestionRating = ServerRatings.Silver;
        testServer.latency = 12345;
        testServer.latencyRating = ServerRatings.Gold;
        testServer.hops = 3;
        testServer.note = "Note";
        servers.add(testServer);

        testServer = new VpnServerForList();
        testServer.lastUsed = new Date();
        testServer.countryCode = "de";
        testServer.name = "Test server 20";
        testServer.location = "Berlin";
        testServer.pk = "044ec47420176680816e0406250e7156465e4531f5b26057c9f6297bb0303558c7";
        testServer.congestion = 20;
        testServer.congestionRating = ServerRatings.Gold;
        testServer.latency = 123;
        testServer.latencyRating = ServerRatings.Bronze;
        testServer.hops = 7;
        servers.add(testServer);

        VPNServersPersistentData.getInstance().updateFromDiscovery(servers);

        if (serverSubscription != null) {
            serverSubscription.dispose();
        }

        serverSubscription = Observable.just(servers).flatMap(serversList ->
            VPNServersPersistentData.getInstance().history()
        ).subscribe(r -> {
            ArrayList<VpnServerForList> serversCopy = new ArrayList<>(servers);

            removeSavedData(serversCopy);
            addSavedData(serversCopy);
            sortList(serversCopy);
            adapter.setData(serversCopy, ServerLists.Public);
        });

        recycler.setVisibility(View.VISIBLE);
        loadingAnimation.setVisibility(View.GONE);
        textNoResults.setVisibility(View.GONE);
    }

    private void addSavedData(ArrayList<VpnServerForList> servers) {
        ArrayList<VpnServerForList> remove = new ArrayList();
        for (VpnServerForList server : servers) {
            LocalServerData savedVersion = VPNServersPersistentData.getInstance().getSavedVersion(server.pk);

            if (savedVersion != null) {
                server.customName = savedVersion.customName;
                server.personalNote = savedVersion.personalNote;
                server.inHistory = savedVersion.inHistory;
                server.flag = savedVersion.flag;
                server.enteredManually = savedVersion.enteredManually;
                server.usedWithPassword = savedVersion.usedWithPassword;
            }

            if (server.flag == ServerFlags.Blocked) {
                remove.add(server);
            }
        }

        servers.removeAll(remove);
    }

    private void removeSavedData(ArrayList<VpnServerForList> servers) {
        for (VpnServerForList server : servers) {
            server.customName = null;
            server.personalNote = null;
            server.inHistory = false;
            server.flag = ServerFlags.None;
            server.enteredManually = false;
            server.usedWithPassword = false;
        }
    }

    private void sortList(ArrayList<VpnServerForList> servers) {
        if (listType != ServerLists.History) {
            Comparator<VpnServerForList> comparator = (a, b) -> {
                int response = a.countryCode.compareTo(b.countryCode);

                if (response == 0) {
                    response = getServerName(a).compareTo(getServerName(b));
                }

                return response;
            };

            Collections.sort(servers, comparator);
        } else {
            Comparator<VpnServerForList> comparator = (a, b) -> (int)((b.lastUsed.getTime() - a.lastUsed.getTime()) / 1000);
            Collections.sort(servers, comparator);
        }
    }

    private String getServerName(VpnServerForList server) {
        if ((server.name == null || server.name.trim().equals("")) && (server.customName == null || server.customName.trim().equals(""))) {
            return "";
        } else if (server.name != null && !server.name.trim().equals("") && (server.customName == null || server.customName.trim().equals(""))) {
            return server.name;
        } else if (server.customName != null && !server.customName.trim().equals("") && (server.name == null || server.name.trim().equals(""))) {
            return server.customName;
        }

        return server.customName + " - " + server.name;
    }
}