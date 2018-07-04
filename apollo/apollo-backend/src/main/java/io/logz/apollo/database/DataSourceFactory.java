package io.logz.apollo.database;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import io.logz.apollo.configuration.DatabaseConfiguration;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.sql.DataSource;
import java.util.Properties;

/**
 * Created by roiravhon on 11/20/16.
 */
public class DataSourceFactory {

    private static final Logger logger = LoggerFactory.getLogger(DataSourceFactory.class);

    private DataSourceFactory() {}

    public static DataSource create(DatabaseConfiguration databaseConfiguration) {
        Properties poolProperties = new Properties();
        poolProperties.setProperty("username", databaseConfiguration.getUser());
        poolProperties.setProperty("dataSourceClassName", com.mysql.jdbc.jdbc2.optional.MysqlDataSource.class.getName());
        poolProperties.setProperty("minimumIdle", String.valueOf(1));
        poolProperties.setProperty("maximumPoolSize", String.valueOf(50));
        poolProperties.setProperty("dataSource.serverName", databaseConfiguration.getHost());
        poolProperties.setProperty("dataSource.port", String.valueOf(databaseConfiguration.getPort()));
        poolProperties.setProperty("dataSource.databaseName", databaseConfiguration.getSchema());
        poolProperties.setProperty("dataSource.useUnicode", String.valueOf(true));
        poolProperties.setProperty("dataSource.characterEncoding", "UTF-8");
        poolProperties.setProperty("poolName", "apollo");
        poolProperties.setProperty("connectionTimeout", String.valueOf(5000));
        poolProperties.setProperty("registerMbeans", "true");

        logger.info("Creating connection pool with these parameters: {}", poolProperties.toString());

        poolProperties.setProperty("password", databaseConfiguration.getPassword());
        HikariConfig hikariConfig = new HikariConfig(poolProperties);

        try {
            return new HikariDataSource(hikariConfig);
        } catch (Exception e) {
            throw new RuntimeException("Could not create connection pool, bailing out!", e);
        }
    }
}
