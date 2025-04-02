package bookingclient;

public class ClientMain {
    public static void main(String[] args) {
        int maxRetries = 3;
        if (args.length > 0) {
            try {
                maxRetries = Integer.parseInt(args[0]);
            } catch (NumberFormatException e) {
                System.out.println("Invalid retry argument. Using default of 3.");
            }
        }

        NetworkHandler networkHandler = new NetworkHandler();
        networkHandler.setMaxRetries(maxRetries);
        networkHandler.networkClient();

        Client client = new Client(networkHandler, System.in);
        client.handleRequest();
    }
}
