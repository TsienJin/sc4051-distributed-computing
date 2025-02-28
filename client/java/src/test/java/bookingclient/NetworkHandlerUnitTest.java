package bookingclient;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.SocketTimeoutException;
import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.List;
import java.util.UUID;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class  NetworkHandlerUnitTest {

    @Mock
    private DatagramSocket mockSocket;

    @Mock
    private InetAddress address;

    @InjectMocks
    private NetworkHandler networkHandler;

    @Captor
    private ArgumentCaptor<DatagramPacket> packetCaptor;

    @BeforeEach
    void setUp() throws Exception {
        networkHandler = new NetworkHandler();
        networkHandler.setSocket(mockSocket);
        lenient().when(address.getHostAddress()).thenReturn("100.105.193.66");
    }

    @Test
    void testSendPacketWithAckAndResend_ACKThenResponseReceived() throws IOException {
        // Sample packet to send
        byte[] packet = new byte[]{0x01, 0x02, 0x03};

        // Prepare mock data for ACK and RESPONSE packets
        UUID messageId = UUID.randomUUID();
        int packetNumber = 0;
        int totalPackets = 1;

        // Construct the payload for the RESPONSE packet (UUID + packet number)
        ByteBuffer payload = ByteBuffer.allocate(16 + 1);
        payload.put(PacketMarshaller.UUIDtoByteArray(messageId));
        payload.put((byte) packetNumber);
        byte[] payloadBytes = payload.array();

        // Marshal the ACK packet
        byte[] ackData = PacketMarshaller.marshalPacket(
                (byte) 0x04, // ACK type code
                (byte) 0x00, // packet number 0
                (byte) 0x01, // total packets 1
                false,
                false,
                payloadBytes
        );

        // Marshal the RESPONSE packet
        byte[] responseData = PacketMarshaller.marshalPacket(
                (byte) 0x03, // RESPONSE type code
                (byte) packetNumber,
                (byte) totalPackets,
                false,
                false,
                payloadBytes
        );

        // Mock the socket's receive to return ACK first, then RESPONSE
        lenient().doAnswer(invocation -> {
            DatagramPacket dp = invocation.getArgument(0);
            dp.setData(ackData);
            dp.setLength(ackData.length);
            dp.setAddress(address);
            dp.setPort(NetworkHandler.UDP_PORT);
            return null;
        }).doAnswer(invocation -> {
            DatagramPacket dp = invocation.getArgument(0);
            dp.setData(responseData);
            dp.setLength(responseData.length);
            dp.setAddress(address);
            dp.setPort(NetworkHandler.UDP_PORT);
            return null;
        }).when(mockSocket).receive(any(DatagramPacket.class));

        // Call the method under test
        List<Packet> response = networkHandler.sendPacketWithAckAndResend(packet);

        // Verify the response
        assertNotNull(response);
        assertEquals(1, response.size());
        Packet receivedPacket = response.get(0);
        assertEquals(0x03, receivedPacket.messageType()); // RESPONSE type code
        assertEquals(packetNumber, receivedPacket.packetNumber());
        assertEquals(totalPackets, receivedPacket.totalPackets());

        // Verify that the ACK was sent
        verify(mockSocket, times(2)).send(packetCaptor.capture()); // Initial send and ACK send

        // The second send is the ACK for the RESPONSE
        DatagramPacket ackPacket = packetCaptor.getAllValues().get(1);
        byte[] sentAckData = Arrays.copyOf(ackPacket.getData(), ackPacket.getLength());

        // Construct the expected ACK data
        ByteBuffer expectedAckPayload = ByteBuffer.allocate(16 + 1);
        expectedAckPayload.put(PacketMarshaller.UUIDtoByteArray(messageId));
        expectedAckPayload.put((byte) packetNumber);
        byte[] expectedCombinedPayload = expectedAckPayload.array();

        byte[] expectedAckData = PacketMarshaller.marshalPacket(
                (byte) 0x04, // ACK type code
                (byte) 0x00, // packet number 0
                (byte) 0x01, // total packets 1
                false,
                false,
                expectedCombinedPayload
        );

    }
    @Test
    void testSendPacketWithAckAndResend_ACKDroppedThenRetrySucceeds() throws IOException {
        // Sample packet to send
        byte[] packet = new byte[]{0x01, 0x02, 0x03};

        // Prepare mock data for ACK and RESPONSE packets
        UUID messageId = UUID.randomUUID();
        int packetNumber = 0;
        int totalPackets = 1;

        // Construct the payload for the packets (UUID + packet number)
        ByteBuffer payload = ByteBuffer.allocate(16 + 1);
        payload.put(PacketMarshaller.UUIDtoByteArray(messageId));
        payload.put((byte) packetNumber);
        byte[] payloadBytes = payload.array();

        // Marshal the ACK packet
        byte[] ackData = PacketMarshaller.marshalPacket(
                (byte) 0x04, // ACK type code
                (byte) 0x00, // packet number 0
                (byte) 0x01, // total packets 1
                false,
                false,
                payloadBytes
        );

        // Marshal the RESPONSE packet
        byte[] responseData = PacketMarshaller.marshalPacket(
                (byte) 0x03, // RESPONSE type code
                (byte) packetNumber,
                (byte) totalPackets,
                false,
                false,
                payloadBytes
        );

        // Mock the socket's receive behavior:
        // 1st attempt: Socket times out (ACK dropped)
        // 2nd attempt: ACK received, then RESPONSE received
        lenient().doThrow(new SocketTimeoutException("Socket timed out"))
                .doAnswer(invocation -> {
                    DatagramPacket dp = invocation.getArgument(0);
                    dp.setData(ackData);
                    dp.setLength(ackData.length);
                    dp.setAddress(address);
                    dp.setPort(NetworkHandler.UDP_PORT);
                    return null;
                }).doAnswer(invocation -> {
                    DatagramPacket dp = invocation.getArgument(0);
                    dp.setData(responseData);
                    dp.setLength(responseData.length);
                    dp.setAddress(address);
                    dp.setPort(NetworkHandler.UDP_PORT);
                    return null;
                }).when(mockSocket).receive(any(DatagramPacket.class));

        // Call the method under test
        List<Packet> response = networkHandler.sendPacketWithAckAndResend(packet);

        // Verify the response
        assertNotNull(response);
        assertEquals(1, response.size());
        Packet receivedPacket = response.get(0);
        assertEquals(0x03, receivedPacket.messageType()); // RESPONSE type code
        assertEquals(packetNumber, receivedPacket.packetNumber());
        assertEquals(totalPackets, receivedPacket.totalPackets());

        // Verify that the packet was sent twice due to missing ACK
        verify(mockSocket, times(3)).send(packetCaptor.capture());
        // Verify sequence: Initial send, resend after timeout, ACK for response

        // The first packet sent should be the original request
        DatagramPacket firstSentPacket = packetCaptor.getAllValues().get(0);
        byte[] firstSentData = Arrays.copyOf(firstSentPacket.getData(), firstSentPacket.getLength());
        assertArrayEquals(packet, firstSentData);

        // The second packet sent should be the original request again (after timeout)
        DatagramPacket resendPacket = packetCaptor.getAllValues().get(1);
        byte[] resendData = Arrays.copyOf(resendPacket.getData(), resendPacket.getLength());
        assertArrayEquals(packet, resendData);

        // The third packet sent should be the ACK for the received RESPONSE
        DatagramPacket ackPacket = packetCaptor.getAllValues().get(2);
        byte[] sentAckData = Arrays.copyOf(ackPacket.getData(), ackPacket.getLength());

        // Verify that the ACK packet has the correct message type
        Packet sentAckPacket = PacketUnmarshaller.unmarshalResponse(sentAckData);
        assertEquals(0x04, sentAckPacket.messageType()); // ACK type code
    }
    @Test
    void testSendPacketWithAckAndResend_ACKReceivedButResponseDropped() throws IOException {
        // Sample packet to send
        byte[] packet = new byte[]{0x01, 0x02, 0x03};

        // Prepare mock data for ACK and RESPONSE packets
        UUID messageId = UUID.randomUUID();
        int packetNumber = 0;
        int totalPackets = 1;

        // Construct the payload for the RESPONSE packet (UUID + packet number)
        ByteBuffer payload = ByteBuffer.allocate(16 + 1);
        payload.put(PacketMarshaller.UUIDtoByteArray(messageId));
        payload.put((byte) packetNumber);
        byte[] payloadBytes = payload.array();

        // Marshal the ACK packet
        byte[] ackData = PacketMarshaller.marshalPacket(
                (byte) 0x04, // ACK type code
                (byte) 0x00, // packet number 0
                (byte) 0x01, // total packets 1
                false,
                false,
                payloadBytes
        );

        // Marshal the RESPONSE packet
        byte[] responseData = PacketMarshaller.marshalPacket(
                (byte) 0x03, // RESPONSE type code
                (byte) packetNumber,
                (byte) totalPackets,
                false,
                false,
                payloadBytes
        );

        // Mock the socket's receive behavior:
        // 1st attempt: ACK received, RESPONSE is dropped
        // 2nd attempt: ACK and RESPONSE received
        lenient().doAnswer(invocation -> {
                    DatagramPacket dp = invocation.getArgument(0);
                    dp.setData(ackData);
                    dp.setLength(ackData.length);
                    dp.setAddress(address);
                    dp.setPort(NetworkHandler.UDP_PORT);
                    return null;
                }).doThrow(new SocketTimeoutException("Socket timed out"))
                .doAnswer(invocation -> {
                    DatagramPacket dp = invocation.getArgument(0);
                    dp.setData(ackData);
                    dp.setLength(ackData.length);
                    dp.setAddress(address);
                    dp.setPort(NetworkHandler.UDP_PORT);
                    return null;
                }).doAnswer(invocation -> {
                    DatagramPacket dp = invocation.getArgument(0);
                    dp.setData(responseData);
                    dp.setLength(responseData.length);
                    dp.setAddress(address);
                    dp.setPort(NetworkHandler.UDP_PORT);
                    return null;
                }).when(mockSocket).receive(any(DatagramPacket.class));

        // Call the method under test
        List<Packet> response = networkHandler.sendPacketWithAckAndResend(packet);

        // Verify the response
        assertNotNull(response);
        assertEquals(1, response.size());
        Packet receivedPacket = response.get(0);
        assertEquals(0x03, receivedPacket.messageType()); // RESPONSE type code
        assertEquals(packetNumber, receivedPacket.packetNumber());
        assertEquals(totalPackets, receivedPacket.totalPackets());

        // Verify that the request was sent twice due to missing RESPONSE
        verify(mockSocket, times(3)).send(packetCaptor.capture()); // Initial send, resend, ACK send

        // Verify that the last sent packet was the ACK for the RESPONSE
        DatagramPacket ackPacket = packetCaptor.getAllValues().get(2);
        byte[] sentAckData = Arrays.copyOf(ackPacket.getData(), ackPacket.getLength());

        // Construct the expected ACK data
        ByteBuffer expectedAckPayload = ByteBuffer.allocate(16 + 1);
        expectedAckPayload.put(PacketMarshaller.UUIDtoByteArray(messageId));
        expectedAckPayload.put((byte) packetNumber);
        byte[] expectedCombinedPayload = expectedAckPayload.array();

        byte[] expectedAckData = PacketMarshaller.marshalPacket(
                (byte) 0x04, // ACK type code
                (byte) 0x00, // packet number 0
                (byte) 0x01, // total packets 1
                false,
                false,
                expectedCombinedPayload
        );
    }

}