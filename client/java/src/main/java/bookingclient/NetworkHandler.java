package bookingclient;
import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.Arrays;

public class NetworkHandler {
    private static final String HOST = "100.105.193.66";
    private static final int UDP_PORT = 8765;
    private static final int TIMEOUT_MS = 2000;
    private static final int MAX_RETRIES = 1;
    private static DatagramSocket socket;
    private static InetAddress address;

    public void networkClient(){
        try {
            socket = new DatagramSocket();
            address = InetAddress.getByName("100.105.193.66");
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public byte[] sendPacketWithAck(byte[] packet) throws IOException {

        DatagramPacket DatagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
        byte[] ackBuffer = new byte[1024];
        DatagramPacket ackPacket = new DatagramPacket(ackBuffer, ackBuffer.length);

        int retries = 0;
        while (retries < MAX_RETRIES) {
            try {
                System.out.println(bytesToHex(DatagramPacket.getData()));
                socket.send(DatagramPacket);
                socket.receive(ackPacket);
                System.out.println(ackPacket);
                return ackPacket.getData();
            } catch (IOException e) {
                retries++;
                System.out.println("Failed to send packet, retrying...");
            }
        }

        throw new IOException("Failed to send packet after " + MAX_RETRIES + " retry attempts.");

    }

    // Helper method to convert byte array to hex string for easier debugging
    public static String bytesToHex(byte[] byteArray) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : byteArray) {
            hexString.append(String.format("%02X", b));
        }
        return hexString.toString();
    }


}
