package io.logz.apollo.kubernetes;

import com.google.common.annotations.VisibleForTesting;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.logz.apollo.dao.DeploymentDao;
import io.logz.apollo.dao.ServiceDao;
import io.logz.apollo.models.Environment;
import io.logz.apollo.notifications.ApolloNotifications;

import javax.inject.Inject;
import javax.inject.Singleton;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import static java.util.Objects.requireNonNull;

@Singleton
public class KubernetesHandlerStore {

    private final ApolloToKubernetesStore apolloToKubernetesStore;
    private final ApolloNotifications apolloNotifications;
    private final Map<Integer, KubernetesHandler> kubernetesHandlerMap;

    @Inject
    public KubernetesHandlerStore(ApolloToKubernetesStore apolloToKubernetesStore, ApolloNotifications apolloNotifications) {
        this.apolloToKubernetesStore = requireNonNull(apolloToKubernetesStore);
        this.apolloNotifications = requireNonNull(apolloNotifications);
        this.kubernetesHandlerMap = new ConcurrentHashMap<>();
    }

    public KubernetesHandler getOrCreateKubernetesHandler(Environment environment) {
        return kubernetesHandlerMap.computeIfAbsent(environment.getId(),
                key -> new KubernetesHandler(apolloToKubernetesStore, environment, apolloNotifications));
    }

    @VisibleForTesting
    public KubernetesHandler getOrCreateKubernetesHandlerWithSpecificClient(Environment environment, KubernetesClient kubernetesClient) {
        return kubernetesHandlerMap.computeIfAbsent(environment.getId(),
                key -> new KubernetesHandler(apolloToKubernetesStore, kubernetesClient, environment, apolloNotifications));
    }
}
