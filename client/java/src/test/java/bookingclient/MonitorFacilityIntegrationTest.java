package bookingclient;

import org.junit.jupiter.api.Test;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.util.Scanner;

public class MonitorFacilityIntegrationTest {
    @Test
    public void testMultipleClients() throws InterruptedException {
        // Start the first client in a separate thread that will book a facility
        Thread client1 = new Thread(() -> {
            try {
                runBookingClient("Client 1 - Booking");
            } catch (IOException e) {
                e.printStackTrace();
            }
        });

        // Start the second client in a separate thread that will monitor the same facility
        Thread client2 = new Thread(() -> {
            try {
                runMonitoringClient("Client 2 - Monitoring");
            } catch (IOException e) {
                e.printStackTrace();
            }
        });

        // Start both threads
        client1.start();
//        client2.start();

        // Wait for both clients to finish their tasks
        client1.join();
        //client2.join();
    }

    private void runBookingClient(String clientName) throws IOException {
        String facilityName = "One";
        int startTime = 4833210;
        int endTime = 483311;
        String bookingDetails = facilityName + "\n" + startTime + "\n" + endTime +"\n" +"9\n";

        // Execute client logic
        NetworkHandler networkHandler = new NetworkHandler();
        networkHandler.networkClient();
        Client client = new Client(networkHandler,new ByteArrayInputStream(bookingDetails.getBytes()));
        client.setState(new BookFacilityState());
        client.handleRequest();
        return;
    }

    private void runMonitoringClient(String clientName) throws IOException {
        String facilityName = "Conference Room A";
        int ttl = 60;
        String monitoringDetails = facilityName + "\n" + ttl + "\n";


        // Execute client logic
        NetworkHandler networkHandler = new NetworkHandler();
        networkHandler.networkClient();
        Client client = new Client(networkHandler,new ByteArrayInputStream(monitoringDetails.getBytes()));
        client.setState(new MonitorFacilityState());
        client.handleRequest();
    }
}
