package bookingclient;

import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteBuffer;
import java.time.LocalTime;
import java.util.List;

interface ClientState {
    void handleRequest(Client client);
}
class MenuState implements ClientState{
    //1,2,3,4 are MANDATORY. 5,6,7,8 not MANDATORY
    @Override
    public void handleRequest(Client client) {
        System.out.println("1. Query Facility Availability");
        System.out.println("2. Book Facility");
        System.out.println("3. Modify Booking");
        System.out.println("4. Monitor Facility");
        System.out.println("5. Create Facility");
        System.out.println("6. Delete Facility");
        System.out.println("7. List Facilities");
        System.out.println("8. Delete Booking");
        System.out.println("9. Exit");

        int choice = client.getUserInputUtils().getIntInput("Choose an option:");
        switch (choice) {
            case 1:
                client.setState(new QueryFacilityState());
                break;
            case 2:
                client.setState(new BookFacilityState());
                break;
            case 4:
                client.setState(new MonitorFacilityState());
                break;
            case 5:
                client.setState(new CreateFacilityState());
                break;
            case 6:
                client.setState(new DeleteFacilityState());
                break;
            case 8:
                client.setState(new DeleteBookingState());
                break;
            case 9:
                System.out.println("Exiting system...");
                System.exit(0);
            default:
                System.out.println("Invalid option. Please try again.");
        }
        client.handleRequest();
    }

}
class MonitorFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = client.getUserInputUtils().getStringInput("Monitored Facility Name:");
        int numberOfSeconds = client.getUserInputUtils().getIntInput("number Of Seconds: ");
        byte[] packet = PacketMarshaller.marshalMonitorFacility(facility, numberOfSeconds);
        try {
            List<Packet> response = client.getNetworkHandler().sendMonitorFacilityPacket(packet, numberOfSeconds);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
        client.setState(new MenuState());
        client.handleRequest();
    }
    }
class CreateFacilityState implements ClientState {
    @Override
    public void handleRequest(Client client) {
        // Get user input for facility name
        String facility = client.getUserInputUtils().getStringInput("Create Facility Name:");
        System.out.println("Creating Facility Name: " + facility);
        byte[] packet = PacketMarshaller.marshalCreateFacilityRequest(facility);
        try {
            List<Packet> response = client.getNetworkHandler().sendPacketWithAckAndResend(packet);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();
    }
}

class DeleteFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = client.getUserInputUtils().getStringInput("Delete Facility Name:");
        System.out.println("Deleting Facility Name " + facility);
        byte[] packet = PacketMarshaller.marshalDeleteFacilityRequest(facility);
        try {
            List<Packet> response = client.getNetworkHandler().sendPacketWithAckAndResend(packet);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();
    }
}

class QueryFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = client.getUserInputUtils().getStringInput("Query Facility Name:");
        int numberOfDays = client.getUserInputUtils().getIntInput("Number of Days:");

        System.out.println("Querying Facility Name " + facility);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        byte[] packet = PacketMarshaller.marshalQueryFacilityRequest(facility, numberOfDays);  // Marshal the facility data
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            List<Packet> response = client.getNetworkHandler().sendPacketWithAckAndResend(packet);
            String payload = PacketMarshaller.bytesToHex(response.get(0).getPayload());
            System.out.println("payload: " + payload);

            // first 36 hex are message ID and status Code
            String output = payload.substring(36);
            System.out.println("output: " + output);

            // TODO: when error code is 400 (facility does not exist),
            // no error msg gets printed out in PacketUnmarshaller.unmarshalResponse()
            int status = response.get(0).getStatus();
            if (status == 200) {
                System.out.println("Printing out availability");
                for (char c : output.toCharArray()) {
                    // Convert the character to a number
                    int num = Character.getNumericValue(c);
                    // Print the number in binary
                    System.out.print(String.format("%4s", Integer.toBinaryString(num)).replace(' ', '0') + " ");
                }
                System.out.println();
            }
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
//        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();

        client.setState(new MenuState());
        client.handleRequest();
    }
}

class BookFacilityState implements ClientState{
    @Override
    public void handleRequest(Client client) {
        String facility = client.getUserInputUtils().getStringInput("Query Facility Name:");
        int startTime = client.getUserInputUtils().getIntInput("Enter hours since UNIX (startTime: " + getHoursSinceUnix() + "): ");
        int endTime = client.getUserInputUtils().getIntInput("Enter hours since UNIX (endTime: " + getHoursSinceUnix() + "): ");

        System.out.println("Booking Facility Name " + facility);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        byte[] packet = PacketMarshaller.marshalBookFacilityRequest(facility, startTime, endTime);  // Marshal the facility data
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            List<Packet> response = client.getNetworkHandler().sendPacketWithAckAndResend(packet);

            String payload = PacketMarshaller.bytesToHex(response.get(0).getPayload());
            String bookindId = payload.substring(payload.length() - 4);
            System.out.println("Booking ID: " + bookindId);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
//        marshaller.unmarshalResponse(ackpacket);

        client.setState(new MenuState());
        client.handleRequest();
    }

    public long getHoursSinceUnix() {
        // Get the current time in milliseconds since Unix epoch
        long currentTimeMillis = System.currentTimeMillis();

        // Convert milliseconds to seconds
        long currentTimeSeconds = currentTimeMillis / 1000;

        // Convert seconds to hours
        long hoursSinceUnixEpoch = currentTimeSeconds / 3600;  // 3600 seconds in an hour

        return hoursSinceUnixEpoch;
    }
}

class DeleteBookingState implements ClientState {
    @Override
    public void handleRequest(Client client) {
        int confirmationCode = client.getUserInputUtils().getHexInput("Delete Booking for Confirmation code:");

        System.out.println("Deleting booking with confirmation code: " + confirmationCode);
        // Create PacketMarshaller and NetworkHandler objects (no singleton here, just direct instantiation)
        byte[] packet = PacketMarshaller.marshalDeleteBookingRequest(confirmationCode);  // Marshal the facility data
        // Directly create the NetworkHandler and send the packet
        NetworkHandler networkHandler = new NetworkHandler();  // Direct instantiation
        networkHandler.networkClient();
        try {
            List<Packet> response = client.getNetworkHandler().sendPacketWithAckAndResend(packet);
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
//        marshaller.unmarshalResponse(ackpacket);
        // After processing, return to MenuState
        client.setState(new MenuState());
        client.handleRequest();

        client.setState(new MenuState());
        client.handleRequest();
    }
}


public class Client {
    private ClientState currentState;
    private NetworkHandler networkHandler;
    private UserInputUtils userInputUtils;

    public Client(NetworkHandler networkHandler, InputStream inputStream) {
        this.networkHandler = networkHandler;
        this.currentState = new MenuState();
        this.userInputUtils = new UserInputUtils(inputStream);  // Initialize UserInputUtils with default System.in
    }

    public NetworkHandler getNetworkHandler() {
        return networkHandler;
    }

    public void setState(ClientState state) {
        this.currentState = state;
    }

    public void handleRequest() {
        currentState.handleRequest(this);  // Delegate request handling to the current state
    }

    // Getter method for UserInputUtils to be used by ClientState classes
    public UserInputUtils getUserInputUtils() {
        return userInputUtils;
    }
}
