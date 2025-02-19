package bookingclient;

import java.util.Scanner;

public class ClientMain {
    public static void main(String[] args) {
        NetworkHandler networkHandler = new NetworkHandler();
        networkHandler.networkClient();
        Client client = new Client(networkHandler,System.in);
        client.handleRequest();
    }
}
