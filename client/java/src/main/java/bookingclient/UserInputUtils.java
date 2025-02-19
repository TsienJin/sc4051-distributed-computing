package bookingclient;

import java.io.InputStream;
import java.util.Scanner;

public class UserInputUtils {

    private Scanner scanner; // Instance variable to hold the Scanner

    // Constructor to initialize with a specific input stream
    public UserInputUtils(InputStream inputStream) {
        this.scanner = new Scanner(inputStream); // Create a new Scanner for the given stream
    }

    // Method to get a valid String input from the user
    public String getStringInput(String prompt) {
        System.out.println(prompt); // Display the prompt to the user
        return scanner.nextLine().trim(); // Read and trim any surrounding spaces from the input
    }

    // Method to get a valid Integer input from the user
    public int getIntInput(String prompt) {
        System.out.println(prompt); // Display the prompt to the user
        int number = scanner.nextInt();
        scanner.nextLine();  // Consume the newline character
        return number;
    }

    // Method to get a valid Hexadecimal input from the user
    public int getHexInput(String prompt) {
        System.out.println(prompt); // Display the prompt to the user
        int number = scanner.nextInt(16); // Read a hexadecimal input
        scanner.nextLine(); // Consume the newline character
        return number;
    }
}
