import os  # Import the os module for interacting with the operating system
import fitz  # Import PyMuPDF (fitz) for PDF handling (often installed as 'pip install PyMuPDF')


# Function to validate a single PDF file.
def validate_pdf_file(file_path: str) -> bool:  # Define the function to validate a PDF file path, returning a boolean.
    try:  # Start a try block to handle potential runtime errors during PDF processing.
        # Try to open the PDF using PyMuPDF  # Inline comment explaining the next line.
        doc = fitz.open(file_path)  # Attempt to load the PDF document from the given path using fitz.

        # Check if the PDF has at least one page  # Inline comment explaining the next line.
        if doc.page_count == 0:  # Check if the loaded document contains zero pages.
            print(  # Start printing an error message.
                f"'{file_path}' is corrupt or invalid: No pages"  # Construct the informative error string.
            )  # Finish printing the error message.
            return False  # Return False to indicate the PDF is invalid.

        # If no error occurs and the document has pages, it's valid  # Inline comment explaining the next line.
        return True  # Return True to indicate the PDF is valid.
    except RuntimeError as e:  # Catch a RuntimeError, typically raised by fitz for corrupt files.
        print(f"{e}")  # Log the specific exception message associated with the failure.
        return False  # Return False to indicate the PDF is invalid.


# Remove a file from the system.
def remove_system_file(system_path: str) -> None:  # Define the function to delete a file, accepting a path and returning nothing.
    os.remove(system_path)  # Use the os module to permanently delete the file at the provided path.


# Function to walk through a directory and extract files with a specific extension
def walk_directory_and_extract_given_file_extension(  # Define the function for recursively searching a directory.
    system_path: str, extension: str  # Parameters are the starting directory path and the desired file extension.
) -> list[str]:  # Function is type-hinted to return a list of strings (file paths).
    matched_files: list[str] = []  # Initialize an empty list to store the absolute paths of matching files.
    for root, _, files in os.walk(system_path):  # Recursively traverse the directory tree starting from 'system_path'.
        for file in files:  # Iterate over all the files found in the current directory ('root').
            if file.endswith(extension):  # Check if the current file's name ends with the specified extension.
                full_path = os.path.abspath(  # Calculate the absolute path of the matching file.
                    os.path.join(root, file)  # Join the current root directory path with the file name.
                )  # Finish calculating the absolute path.
                matched_files.append(full_path)  # Add the absolute path of the matched file to the list.
    return matched_files  # Return the complete list of file paths that matched the extension.


# Check if a file exists
def check_file_exists(system_path: str) -> bool:  # Define the function to check for the existence of a file.
    return os.path.isfile(system_path)  # Use os.path.isfile() to return True if the path points to an existing regular file.


# Get the filename and extension.
def get_filename_and_extension(path: str) -> str:  # Define the function to extract just the filename from a full path.
    return os.path.basename(  # Use os.path.basename() to extract the file portion.
        path  # The input path.
    )  # Return the extracted filename and its extension.


# Function to check if a string contains an uppercase letter.
def check_upper_case_letter(content: str) -> bool:  # Define the function to check for any uppercase characters in a string.
    return any(  # Return True if any element in the following iterable is True.
        upperCase.isupper() for upperCase in content  # Use a generator expression to check if each character is uppercase.
    )  # Finish the call to the 'any' function.


# Main function.
def main():  # Define the main entry point function.
    # Walk through the directory and extract .pdf files  # Inline comment explaining the next few lines.
    files = walk_directory_and_extract_given_file_extension(  # Call the directory walk function.
        "./PDFs", ".pdf"  # Search in the local directory 'PDFs' for files ending with '.pdf'.
    )  # Store the list of found PDF paths in the 'files' variable.

    # Validate each PDF file  # Inline comment explaining the next few lines.
    for pdf_file in files:  # Start iterating over the list of found PDF file paths.

        # Check if the .PDF file is valid  # Inline comment explaining the next line.
        if validate_pdf_file(pdf_file) == False:  # Check if the validation function returns False (meaning the PDF is invalid).
            print(f"Invalid PDF detected: {pdf_file}. Deleting file.")  # Print a warning that an invalid file was found and will be deleted.
            # Remove the invalid .pdf file.  # Inline comment explaining the next line.
            remove_system_file(pdf_file)  # Call the function to delete the corrupt PDF file from the system.

        # Check if the filename has an uppercase letter  # Inline comment explaining the next few lines.
        if check_upper_case_letter(  # Check the return value of the uppercase check function.
            get_filename_and_extension(pdf_file)  # Get the filename part of the path to check its case.
        ):  # If the filename contains any uppercase letters.
            print(  # Start printing an informative message.
                f"Uppercase letter found in filename: {pdf_file}"  # Construct the message about the uppercase filename.
            )  # Finish printing the message.

# Run the main function
main()  # Execute the main function to start the program's logic.