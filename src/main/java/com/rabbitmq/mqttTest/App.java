package com.rabbitmq.mqttTest;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Random;
import java.util.UUID;
import java.util.concurrent.CountDownLatch;

import org.eclipse.paho.client.mqttv3.MqttClient;
import org.eclipse.paho.client.mqttv3.MqttConnectOptions;
import org.eclipse.paho.client.mqttv3.MqttMessage;
import org.eclipse.paho.client.mqttv3.persist.MemoryPersistence;

public class App 
{
    public static int            NB_CLIENTS;
    public static int            NB_THREADS;

    public static boolean        SEND_MSG           = true;
    public static long           MSG_TO_SEND_PER_CLI;
    public static long           PAUSE_BTW_SEND     = 200;
    public static byte[]         MSG_CONTENT        = "some data".getBytes();

    public static boolean        JOIN_BEFORE_CON    = true,
                                 JOIN_BEFORE_DISCO  = true;
    public static CountDownLatch START_SIGNAL       = new CountDownLatch(1),
                                 STOP_SIGNAL        = new CountDownLatch(NB_THREADS);
    public static int            SLEEP_BEFORE_DISCO = 1000000;

    public static int            MQTT_PING_INTERVAL = 10;
    
    public static List<String>   NODE_IPS           = Arrays.asList("127.0.0.1:1883", "127.0.0.1:1884");
    public static String         RABBIT_USERNAME    = "guest";
    public static String         RABBIT_PASSWORD    = "guest";

    private static final Random random = new Random();

    public static void main( String[] args )
    {
        NB_CLIENTS = Integer.parseInt(args[0]);
        NB_THREADS = Integer.parseInt(args[1]);
        MSG_TO_SEND_PER_CLI = Integer.parseInt(args[2]);

        for (int i = 0; i < NB_THREADS; i++) {
            try {
                new Thread(new Worker()).start();
            }
            catch (Exception e) {
                e.printStackTrace();
            }
        }
        START_SIGNAL.countDown();

        if (JOIN_BEFORE_DISCO) {
            try {
                STOP_SIGNAL.await();
            }
            catch (InterruptedException e) {
                e.printStackTrace();
            }
        }
    }

    private static String getNodeIpPort(int i) {
        return NODE_IPS.get(i % NODE_IPS.size());
    }

    static class Worker implements Runnable {

        List<MqttClient>   clients  = new ArrayList<>();
        MqttConnectOptions connOpts = new MqttConnectOptions();

        public Worker() throws Exception {

            connOpts.setUserName(RABBIT_USERNAME);
            connOpts.setPassword(RABBIT_PASSWORD.toCharArray());
            connOpts.setKeepAliveInterval(MQTT_PING_INTERVAL);
            connOpts.setCleanSession(true);

            for (int i = 0; i < (NB_CLIENTS / NB_THREADS); i++) {
                // MqttClient client = new MqttClient("tcp://" + getNodeIpPort(i) + ":1883",
                MqttClient client = new MqttClient("tcp://" + getNodeIpPort(i),
                        UUID.randomUUID().toString().substring(0, 7), new MemoryPersistence());
                clients.add(client);
            }
        }

        @Override
        public void run() {
            try {
                if (JOIN_BEFORE_CON) {
                    START_SIGNAL.await();
                }

                for (MqttClient client : clients) {
                    client.connect(connOpts);
                    client.subscribe("device/" + client.getClientId() + "/*");
                }

                if (SEND_MSG) {
                    for (int i = 0; i < MSG_TO_SEND_PER_CLI; i++) {
                        for (MqttClient client : clients) {
                            client.publish("device/" + client.getClientId() + "/fake",
                                    new MqttMessage(MSG_CONTENT));
                        }
                        Thread.sleep(PAUSE_BTW_SEND * random.nextInt(5));
                    }
                }

                Thread.sleep(SLEEP_BEFORE_DISCO);

                STOP_SIGNAL.countDown();

            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }
}
